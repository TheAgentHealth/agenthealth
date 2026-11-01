package core

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// CI reports intentionally contain only engine status and structural identities.
// Arbitrary endpoint names, messages and topology never become CI commands.
type junitFailure struct {
	Message string `xml:"message,attr"`
}
type junitCase struct {
	Name    string        `xml:"name,attr"`
	Class   string        `xml:"classname,attr"`
	Failure *junitFailure `xml:"failure,omitempty"`
}
type junitSuite struct {
	XMLName  xml.Name    `xml:"testsuite"`
	Name     string      `xml:"name,attr"`
	Tests    int         `xml:"tests,attr"`
	Failures int         `xml:"failures,attr"`
	Cases    []junitCase `xml:"testcase"`
}

// walkCI keeps dependency occurrences distinct: shared nodes can have different
// relationship policies. The original result remains the authoritative evidence.
func walkCI(results []Result, visit func(string, Result)) {
	var walk func(string, Result)
	walk = func(path string, r Result) {
		visit(path, r)
		for i, dep := range r.Dependencies {
			walk(fmt.Sprintf("%s/dependency-%d", path, i+1), dep)
		}
	}
	for i, r := range results {
		walk(fmt.Sprintf("target-%d", i+1), r)
	}
}

// WriteJUnit emits one case per target/dependency occurrence. Every non-HEALTHY
// status fails the case; no inconclusive health result is treated as skipped.
func WriteJUnit(w io.Writer, results []Result) error {
	if err := validateResults(results); err != nil {
		return err
	}
	suite := junitSuite{Name: "AgentHealth"}
	walkCI(results, func(path string, r Result) {
		c := junitCase{Name: path, Class: "agenthealth." + r.Target.Type}
		if r.Status != Healthy {
			c.Failure = &junitFailure{Message: string(r.Status)}
			suite.Failures++
		}
		suite.Cases = append(suite.Cases, c)
	})
	suite.Tests = len(suite.Cases)
	var buffer bytes.Buffer
	buffer.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buffer)
	encoder.Indent("", "  ")
	if err := encoder.Encode(suite); err != nil {
		return err
	}
	buffer.WriteByte('\n')
	_, err := w.Write(buffer.Bytes())
	return err
}

// WriteGitHub emits safe GitHub workflow annotations. Names and diagnostic text
// are omitted because they can contain secrets or workflow command syntax.
func WriteGitHub(w io.Writer, results []Result) error {
	if err := validateResults(results); err != nil {
		return err
	}
	var buffer strings.Builder
	walkCI(results, func(path string, r Result) {
		level := "error"
		if r.Status == Healthy {
			level = "notice"
		}
		fmt.Fprintf(&buffer, "::%s title=AgentHealth::%s (%s): %s\n", level, path, r.Target.Type, r.Status)
	})
	_, err := io.WriteString(w, buffer.String())
	return err
}
