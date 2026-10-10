// Package agenthealthref provides a reference application contract, not an SDK.
package agenthealthref

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

// JSON escaping can expand each text byte sixfold; framing and downstream
// identifiers need additional room. Decoded fields retain their contract limits.
const maxTaskBodyBytes = 512 * 1024

type Document struct {
	Version      string   `json:"version"`
	Name         string   `json:"name"`
	Live         bool     `json:"live"`
	Ready        bool     `json:"ready"`
	Capabilities []string `json:"capabilities"`
	Dependencies []string `json:"dependencies"`
}
type Task struct {
	Safe       bool   `json:"safe"`
	Text       string `json:"text"`
	Downstream string `json:"downstream,omitempty"`
}
type Outcome struct {
	Completed  bool   `json:"completed"`
	Success    bool   `json:"success"`
	Downstream string `json:"downstream,omitempty"`
}

// Handler serves HEAD/GET and an optional explicitly authorized POST probe.
// Protect this resource with your application's authentication middleware.
// Probe must honor cancellation and perform only bounded, non-destructive work.
func Handler(snapshot func() Document, probe func(context.Context, Task) (Outcome, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		switch r.Method {
		case http.MethodHead:
			w.WriteHeader(200)
		case http.MethodGet:
			doc := snapshot()
			doc.Version = "v1"
			if doc.Capabilities == nil {
				doc.Capabilities = []string{}
			}
			if doc.Dependencies == nil {
				doc.Dependencies = []string{}
			}
			_ = json.NewEncoder(w).Encode(doc)
		case http.MethodPost:
			if probe == nil {
				w.WriteHeader(405)
				return
			}
			var task Task
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxTaskBodyBytes))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&task); err != nil {
				var limit *http.MaxBytesError
				if errors.As(err, &limit) {
					w.WriteHeader(413)
				} else {
					w.WriteHeader(400)
				}
				return
			}
			if !task.Safe || strings.TrimSpace(task.Text) == "" || len(task.Text) > 65536 || utf8.RuneCountInString(task.Downstream) > 1024 || (task.Downstream != "" && strings.TrimSpace(task.Downstream) == "") {
				w.WriteHeader(400)
				return
			}
			var extra any
			if err := decoder.Decode(&extra); err != io.EOF {
				var limit *http.MaxBytesError
				if errors.As(err, &limit) {
					w.WriteHeader(413)
				} else {
					w.WriteHeader(400)
				}
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), time.Second)
			defer cancel()
			result, err := probe(ctx, task)
			if err != nil || ctx.Err() != nil {
				w.WriteHeader(503)
				return
			}
			_ = json.NewEncoder(w).Encode(result)
		default:
			w.Header().Set("Allow", "HEAD, GET, POST")
			w.WriteHeader(405)
		}
	})
}
