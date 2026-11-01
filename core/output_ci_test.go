package core

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"
)

func ciResult(status Status) Result {
	return Result{Target: TargetIdentity{Name: "secret\n::error::injected", Type: "agent"}, Status: status,
		Checks: map[string]CheckResult{"reachability": {Status: status, Message: "password=secret"}}, Dependencies: []Result{}}
}
func TestCIReports(t *testing.T) {
	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			root := ciResult(Healthy)
			root.Dependencies = []Result{ciResult(status), ciResult(status)}
			var out bytes.Buffer
			if err := WriteJUnit(&out, []Result{root}); err != nil {
				t.Fatal(err)
			}
			var suite junitSuite
			if err := xml.Unmarshal(out.Bytes(), &suite); err != nil {
				t.Fatal(err)
			}
			want := 2
			if status == Healthy {
				want = 0
			}
			if suite.Tests != 3 || suite.Failures != want || len(suite.Cases) != 3 {
				t.Fatalf("bad counts: %+v", suite)
			}
			if suite.Cases[1].Name == suite.Cases[2].Name {
				t.Fatal("dependency occurrence lost")
			}
			if strings.Contains(out.String(), "secret") {
				t.Fatal("identity/message leaked")
			}
			out.Reset()
			if err := WriteGitHub(&out, []Result{root}); err != nil {
				t.Fatal(err)
			}
			if strings.Count(out.String(), "\n") != 3 || strings.Contains(out.String(), "secret") {
				t.Fatal(out.String())
			}
			if status != Healthy && !strings.Contains(out.String(), "::error title=AgentHealth::target-1/dependency-1") {
				t.Fatal(out.String())
			}
		})
	}
}

type ciFailWriter struct{}

func (ciFailWriter) Write([]byte) (int, error) { return 0, errors.New("writer failed") }
func TestCIReportsRejectInvalidAndWriterFailures(t *testing.T) {
	for _, write := range []func(io.Writer, []Result) error{WriteJUnit, WriteGitHub} {
		for _, results := range [][]Result{nil, {{}}, {ciResult(Status("INJECTED\n::error::"))}} {
			var out bytes.Buffer
			if write(&out, results) == nil || out.Len() != 0 {
				t.Fatal("invalid results emitted")
			}
		}
		if write(ciFailWriter{}, []Result{ciResult(Healthy)}) == nil {
			t.Fatal("writer error lost")
		}
		cyclic := ciResult(Healthy)
		cyclic.Dependencies = make([]Result, 1)
		cyclic.Dependencies[0] = cyclic
		if write(io.Discard, []Result{cyclic}) == nil {
			t.Fatal("cycle accepted")
		}
	}
}
