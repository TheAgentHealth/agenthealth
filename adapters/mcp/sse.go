package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

// Resumption continues the original request via GET; it never replays POST.
// Byte and reconnect bounds apply across all connections for this response.
func (s *session) readSSE(ctx context.Context, resp *http.Response, expected int) (json.RawMessage, error) {
	remaining := int64(maxBody + 1)
	lastID := ""
	retry := time.Duration(0)
	for attempt := 0; ; attempt++ {
		reader := &io.LimitedReader{R: resp.Body, N: remaining}
		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 4096), maxBody+1)
		var data []byte
		for scanner.Scan() {
			if reader.N == 0 {
				resp.Body.Close()
				return nil, fail(core.Unknown, "mcp_limit")
			}
			line := scanner.Text()
			if line == "" {
				if len(data) > 0 {
					var envelope map[string]json.RawMessage
					if json.Unmarshal(data, &envelope) != nil {
						resp.Body.Close()
						return nil, protocolError()
					}
					if _, hasID := envelope["id"]; hasID {
						resp.Body.Close()
						return decodeResponse(data, expected)
					}
					if string(envelope["jsonrpc"]) != `"2.0"` || !nonemptyString(envelope["method"]) {
						resp.Body.Close()
						return nil, protocolError()
					}
				}
				data = nil
				continue
			}
			field, value, _ := strings.Cut(line, ":")
			value = strings.TrimPrefix(value, " ")
			switch field {
			case "data":
				if len(data) > 0 {
					data = append(data, '\n')
				}
				data = append(data, value...)
			case "id":
				if !strings.ContainsRune(value, '\x00') {
					if len(value) > 4096 || !headerValue(value) {
						resp.Body.Close()
						return nil, protocolError()
					}
					lastID = value
				}
			case "retry":
				if value != "" && strings.Trim(value, "0123456789") == "" {
					n, err := strconv.ParseUint(value, 10, 63)
					if err != nil || n > uint64((1<<63-1)/int64(time.Millisecond)) {
						retry = time.Duration(1<<63 - 1)
					} else {
						retry = time.Duration(n) * time.Millisecond
					}
				}
			}
		}
		remaining = reader.N
		streamErr := scanner.Err()
		resp.Body.Close()
		if remaining == 0 || errors.Is(streamErr, bufio.ErrTooLong) {
			return nil, fail(core.Unknown, "mcp_limit")
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if s.modern || lastID == "" {
			if streamErr != nil {
				return nil, streamErr
			}
			return nil, protocolError()
		}
		if attempt == 3 {
			return nil, fail(core.Unknown, "mcp_limit")
		}
		timer := time.NewTimer(retry)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
		req, err := s.request(ctx, http.MethodGet, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Last-Event-ID", lastID)
		resp, err = s.r.Client.Do(req)
		if err != nil {
			return nil, err
		}
		s.mark()
		media, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			resp.Body.Close()
			return nil, fail(core.Misconfigured, "mcp_auth")
		}
		if resp.StatusCode != http.StatusOK || err != nil || media != "text/event-stream" {
			resp.Body.Close()
			return nil, protocolError()
		}
	}
}
