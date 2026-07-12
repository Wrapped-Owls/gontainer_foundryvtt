package dashboard

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/logstore"
)

// stubLogs implements dashboard.LogReader for testing.
type stubLogs struct {
	lines      []string
	events     []logstore.Event
	next       int
	lastTail   int
	lastCursor int
}

func (s *stubLogs) Logs(n int) []string {
	s.lastTail = n
	return s.lines
}

func (s *stubLogs) Events(cursor int) ([]logstore.Event, int) {
	s.lastCursor = cursor
	return s.events, s.next
}

func serveLogHandlers(t *testing.T, lr *stubLogs) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	registerLogHandlers(mux, lr, slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestGetLogs_passesTail(t *testing.T) {
	lr := &stubLogs{lines: []string{"a", "b"}}
	srv := serveLogHandlers(t, lr)
	resp, err := srv.Client().Get(srv.URL + "/logs?tail=50")
	if err != nil {
		t.Fatalf("get logs: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var got logsResponse
	if err = json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if lr.lastTail != 50 || len(got.Lines) != 2 {
		t.Errorf("unexpected: tail=%d lines=%v", lr.lastTail, got.Lines)
	}
}

func TestGetEvents_passesCursorAndReturnsNext(t *testing.T) {
	lr := &stubLogs{
		events: []logstore.Event{{Kind: "crash", Message: "boom"}},
		next:   7,
	}
	srv := serveLogHandlers(t, lr)
	resp, err := srv.Client().Get(srv.URL + "/events?since=3")
	if err != nil {
		t.Fatalf("get events: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var got eventsResponse
	if err = json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if lr.lastCursor != 3 || got.Next != 7 || len(got.Events) != 1 {
		t.Errorf("unexpected: cursor=%d next=%d events=%v", lr.lastCursor, got.Next, got.Events)
	}
}
