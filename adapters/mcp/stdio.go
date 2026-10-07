package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

type stdioMessage struct {
	data []byte
	err  error
}
type stdioTransport struct {
	cmd      *exec.Cmd
	input    io.WriteCloser
	output   io.ReadCloser
	messages chan stdioMessage
	exited   chan error
	cancel   context.CancelFunc
	writeMu  sync.Mutex
	ignored  map[int]bool
}

func startStdio(ctx context.Context, r core.Request) (*stdioTransport, error) {
	o := r.Target.MCP.Stdio
	processCtx, cancel := context.WithCancel(ctx)
	command, err := stdioCommand(o)
	if err != nil {
		cancel()
		return nil, fail(core.Misconfigured, "mcp_process")
	}
	cmd := exec.CommandContext(processCtx, command, o.Args...)
	cmd.Dir = o.Directory
	cmd.Stderr = io.Discard
	for _, key := range []string{"PATH", "HOME", "USER", "TMPDIR", "TMP", "TEMP", "SYSTEMROOT", "WINDIR"} {
		if v, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+v)
		}
	}
	for key, ref := range o.Env {
		value := r.Environment[ref]
		if value == "" {
			cancel()
			return nil, fail(core.Misconfigured, "mcp_process")
		}
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	configureProcess(cmd)
	cmd.Cancel = func() error { return killProcess(cmd) }
	cmd.WaitDelay = 250 * time.Millisecond
	input, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fail(core.Misconfigured, "mcp_process")
	}
	output, outputWriter, err := os.Pipe()
	if err != nil {
		input.Close()
		cancel()
		return nil, fail(core.Misconfigured, "mcp_process")
	}
	cmd.Stdout = outputWriter
	if err := cmd.Start(); err != nil {
		outputWriter.Close()
		input.Close()
		output.Close()
		cancel()
		return nil, fail(core.Misconfigured, "mcp_process")
	}
	outputWriter.Close()
	p := &stdioTransport{cmd: cmd, input: input, output: output, messages: make(chan stdioMessage, 1), exited: make(chan error, 1), cancel: cancel, ignored: map[int]bool{}}
	go func() {
		scanner := bufio.NewScanner(responseReader{Reader: output, mark: r.MarkResponse})
		scanner.Buffer(make([]byte, 4096), maxBody+1)
		for scanner.Scan() {
			data := append([]byte{}, scanner.Bytes()...)
			select {
			case p.messages <- stdioMessage{data: data}:
			case <-processCtx.Done():
				return
			}
		}
		err := scanner.Err()
		if err == nil {
			err = io.EOF
		}
		select {
		case p.messages <- stdioMessage{err: err}:
		case <-processCtx.Done():
		}
	}()
	go func() { p.exited <- cmd.Wait() }()
	go func() { <-processCtx.Done(); input.Close(); output.Close() }()
	return p, nil
}
func (p *stdioTransport) write(ctx context.Context, data []byte) error {
	done := make(chan error, 1)
	go func() {
		p.writeMu.Lock()
		defer p.writeMu.Unlock()
		_, err := p.input.Write(append(append([]byte{}, data...), '\n'))
		done <- err
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (p *stdioTransport) rpc(ctx context.Context, body []byte, expected int, notification bool, mark func()) (json.RawMessage, error) {
	if err := p.write(ctx, body); err != nil {
		return nil, err
	}
	if notification {
		return nil, nil
	}
	total := 0
	for {
		select {
		case <-ctx.Done():
			p.ignored[expected] = true
			// Best-effort cancellation is bounded independently of a stalled stdin.
			cancelCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			message, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "method": "notifications/cancelled", "params": map[string]any{"requestId": expected}})
			p.write(cancelCtx, message)
			cancel()
			return nil, ctx.Err()
		case message := <-p.messages:
			if message.err != nil {
				if errors.Is(message.err, io.EOF) && total == 0 {
					return nil, fail(core.Unreachable, "mcp_process")
				}
				if errors.Is(message.err, bufio.ErrTooLong) {
					return nil, fail(core.Unknown, "mcp_limit")
				}
				return nil, protocolError()
			}
			mark()
			total += len(message.data) + 1
			if total > maxBody {
				return nil, fail(core.Unknown, "mcp_limit")
			}
			var envelope map[string]json.RawMessage
			if json.Unmarshal(message.data, &envelope) != nil || string(envelope["jsonrpc"]) != `"2.0"` {
				return nil, protocolError()
			}
			if raw, exists := envelope["id"]; exists {
				var id int
				json.Unmarshal(raw, &id)
				if p.ignored[id] {
					delete(p.ignored, id)
					continue
				}
				return decodeResponse(message.data, expected)
			}
			if !nonemptyString(envelope["method"]) {
				return nil, protocolError()
			}
		}
	}
}
func (p *stdioTransport) close() {
	p.input.Close()
	select {
	case <-p.exited:
	case <-time.After(200 * time.Millisecond):
		killProcess(p.cmd)
		select {
		case <-p.exited:
		case <-time.After(250 * time.Millisecond):
		}
	}
	// Kill any descendants that kept the process group alive after the leader exited.
	killProcess(p.cmd)
	p.cancel()
	p.output.Close()
}

type responseReader struct {
	io.Reader
	mark func()
}

func (r responseReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if n > 0 && r.mark != nil {
		r.mark()
	}
	return n, err
}

func stdioCommand(options *core.MCPStdio) (string, error) {
	command := options.Command
	if options.Directory != "" && !filepath.IsAbs(command) && strings.ContainsAny(command, "/\\") {
		path, err := filepath.Abs(filepath.Join(options.Directory, command))
		if err != nil {
			return "", err
		}
		command = path
	}
	return exec.LookPath(command)
}
