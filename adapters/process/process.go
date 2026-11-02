// Package process hosts explicitly authorized executable adapters. A child
// process provides lifecycle separation, not an operating-system sandbox.
package process

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

const maxBytes = 64 << 10

var errProtocol = errors.New("executable adapter failed")

// Manifest is operator-controlled configuration, separate from target YAML.
// Command must be an absolute path; SHA256 pins its bytes, not its dependencies.
type Manifest struct {
	Protocol string        `json:"protocol"`
	Command  string        `json:"command"`
	SHA256   string        `json:"sha256"`
	Metadata core.Metadata `json:"metadata"`
}

type Adapter struct{ manifest Manifest }

// Load validates a bounded manifest without running its executable.
func Load(path string) (*Adapter, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errProtocol
	}
	defer f.Close()
	var m Manifest
	raw, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil || len(raw) > maxBytes {
		return nil, errProtocol
	}
	if err := decode(bytes.NewReader(raw), &m); err != nil {
		return nil, errProtocol
	}
	return New(m)
}

// New validates metadata using the same rules as compiled adapters.
func New(m Manifest) (*Adapter, error) {
	hash, err := hex.DecodeString(m.SHA256)
	if m.Protocol != "v1" || !filepath.IsAbs(m.Command) || err != nil || len(hash) != sha256.Size || hex.EncodeToString(hash) != m.SHA256 {
		return nil, errProtocol
	}
	m.Metadata = clone(m.Metadata)
	a := &Adapter{manifest: m}
	if err := core.NewRegistry().Register(a); err != nil {
		return nil, errProtocol
	}
	return a, nil
}

func clone(m core.Metadata) core.Metadata {
	m.TargetTypes = append([]string(nil), m.TargetTypes...)
	m.Dimensions = append([]string(nil), m.Dimensions...)
	m.ActiveChecks = append([]string(nil), m.ActiveChecks...)
	return m
}
func (a *Adapter) Metadata() core.Metadata { return clone(a.manifest.Metadata) }

// WireRequest intentionally exposes only the selected target's identity,
// endpoint and explicitly configured bearer credential. No graph or environment
// is inherited. Checks requiring richer protocol configuration use compiled adapters.
type WireRequest struct {
	Protocol   string `json:"protocol"`
	Dimension  string `json:"dimension"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Endpoint   string `json:"endpoint"`
	Credential string `json:"credential,omitempty"`
}

// Event is either a response marker or one final normalized status.
// Mark response immediately on receipt of any target response, before reading it.
type Event struct {
	Kind   string      `json:"kind"`
	Status core.Status `json:"status,omitempty"`
}

func decode(r io.Reader, value any) error {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return errProtocol
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errProtocol
	}
	return nil
}

// Check launches one process per dimension; the engine owns aggregation,
// active-check authorization, retries, timing and output diagnostics.
func (a *Adapter) Check(ctx context.Context, r core.Request, dimension string) (obs core.Observation, err error) {
	defer func() {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
	}()
	supported := false
	for _, d := range a.manifest.Metadata.Dimensions {
		if d == dimension {
			supported = true
		}
	}
	if !supported {
		return obs, errProtocol
	}
	file, err := os.Open(a.manifest.Command)
	if err != nil {
		return obs, errProtocol
	}
	stat, err := file.Stat()
	if err != nil || !stat.Mode().IsRegular() {
		file.Close()
		return obs, errProtocol
	}
	hash := sha256.New()
	_, err = io.Copy(hash, file)
	file.Close()
	if err != nil || hex.EncodeToString(hash.Sum(nil)) != a.manifest.SHA256 {
		return obs, errProtocol
	}
	payload, err := json.Marshal(WireRequest{"v1", dimension, r.Target.Name, r.Target.Type, r.Target.Endpoint, r.Credential})
	if err != nil {
		return obs, errProtocol
	}
	cmd := exec.CommandContext(ctx, a.manifest.Command)
	configureProcess(cmd)
	control := processControl{group: func() error { return killProcess(cmd) }, direct: func() error { return cmd.Process.Kill() }}
	cmd.Cancel = control.cancel
	cmd.Env = []string{}
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Stderr = io.Discard
	cmd.WaitDelay = 100 * time.Millisecond
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return obs, errProtocol
	}
	if err := cmd.Start(); err != nil {
		return obs, errProtocol
	}
	// Always terminate/reap a child on malformed or oversized output.
	waited := false
	wait := func() error { control.beginWait(); waited = true; return cmd.Wait() }
	defer func() {
		if !waited {
			_ = control.cancel()
			_ = wait()
		}
	}()
	stop := context.AfterFunc(ctx, func() { _ = stdout.Close() })
	defer stop()
	bounded := &io.LimitedReader{R: stdout, N: maxBytes + 1}
	d := json.NewDecoder(bounded)
	d.DisallowUnknownFields()
	count := 0
	for {
		var event Event
		if err := d.Decode(&event); err != nil {
			return obs, errProtocol
		}
		count++
		if count > 2 || d.InputOffset() > maxBytes {
			return obs, errProtocol
		}
		switch event.Kind {
		case "response":
			if count != 1 || event.Status != "" {
				return obs, errProtocol
			}
			obs.ResponseReceived = true
			if r.MarkResponse != nil {
				r.MarkResponse()
			}
		case "result":
			if !event.Status.Valid() {
				return obs, errProtocol
			}
			obs.Check.Status = event.Status
			var extra any
			if d.Decode(&extra) != io.EOF || bounded.N == 0 || d.InputOffset() > maxBytes {
				return core.Observation{ResponseReceived: obs.ResponseReceived}, errProtocol
			}
			if err := wait(); err != nil {
				return core.Observation{ResponseReceived: obs.ResponseReceived}, errProtocol
			}
			return obs, nil
		default:
			return obs, errProtocol
		}
	}
}

// Keep the package safe for concurrent checks: Adapter has no mutable run state.
var _ core.Adapter = (*Adapter)(nil)

// Group signals are allowed only before reaping can begin. Once Wait starts,
// os.Process.Kill uses Go's lifetime-aware process handle; a raw negative PID
// must never be signaled after a child may have been reaped.
type processControl struct {
	mu            sync.Mutex
	reaping       bool
	group, direct func() error
}

func (p *processControl) cancel() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.reaping {
		return p.direct()
	}
	return p.group()
}
func (p *processControl) beginWait() {
	p.mu.Lock()
	p.reaping = true
	p.mu.Unlock()
}
