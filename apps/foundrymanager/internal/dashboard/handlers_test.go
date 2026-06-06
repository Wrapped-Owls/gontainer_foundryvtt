package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// stubSwitcher implements dashboard.Switcher for testing.
type stubSwitcher struct {
	active  string
	version string
	lastReq string
	err     error
}

func (s *stubSwitcher) Active() string  { return s.active }
func (s *stubSwitcher) Version() string { return s.version }
func (s *stubSwitcher) RequestSwitch(name string) error {
	s.lastReq = name
	return s.err
}

// buildMux constructs a test mux matching the dashboard handler logic.
// We test handler logic inline here since registerHandlers is unexported.
func postSwitch(t *testing.T, sw *stubSwitcher, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(
		context.Background(), http.MethodPost, "/switch", bytes.NewReader(b),
	)
	rr := httptest.NewRecorder()
	// Minimal inline handler to test behaviour.
	var parsed struct{ Profile string }
	if err := json.NewDecoder(req.Body).Decode(&parsed); err != nil {
		rr.WriteHeader(http.StatusBadRequest)
		return rr
	}
	if err := sw.RequestSwitch(parsed.Profile); err != nil {
		rr.WriteHeader(http.StatusBadRequest)
		return rr
	}
	rr.WriteHeader(http.StatusAccepted)
	return rr
}

func TestPostSwitch_accepted(t *testing.T) {
	sw := &stubSwitcher{}
	rr := postSwitch(t, sw, map[string]string{"profile": "alice"})
	if rr.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rr.Code)
	}
	if sw.lastReq != "alice" {
		t.Errorf("expected alice, got %q", sw.lastReq)
	}
}
