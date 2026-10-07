package mcp

import (
	"context"
	"sync"
	"time"
)

// targetSession serializes access even if a canceled adapter call is still
// unwinding. Each target owns its process/session, never another target or run.
type targetSession struct {
	lock    chan struct{}
	session *session
}

func (h *targetSession) acquire(ctx context.Context) bool {
	select {
	case h.lock <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}
func (h *targetSession) release() { <-h.lock }
func sessionState(state *sync.Map) *targetSession {
	value, _ := state.LoadOrStore("mcp-session", &targetSession{lock: make(chan struct{}, 1)})
	return value.(*targetSession)
}
func (Adapter) CloseTarget(ctx context.Context, state *sync.Map) {
	if value, ok := state.Load("mcp-session"); ok {
		h := value.(*targetSession)
		if h.acquire(ctx) {
			defer h.release()
			if h.session != nil {
				h.session.close(ctx)
				h.session = nil
			}
		}
	}
}
func cleanupSession(s *session) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.close(ctx)
}
