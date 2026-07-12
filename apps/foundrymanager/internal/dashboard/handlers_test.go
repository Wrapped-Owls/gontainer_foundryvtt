package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/foundrystatus"
)

// stubSwitcher implements dashboard.Switcher for testing.
type stubSwitcher struct {
	active    string
	version   string
	lastReq   string
	err       error
	status    foundrystatus.Status
	statusErr error
}

func (s *stubSwitcher) Active() string  { return s.active }
func (s *stubSwitcher) Version() string { return s.version }
func (s *stubSwitcher) RequestSwitch(name string) error {
	s.lastReq = name
	return s.err
}

func (s *stubSwitcher) FoundryStatus(_ context.Context) (foundrystatus.Status, error) {
	return s.status, s.statusErr
}

func serveHandlers(t *testing.T, sw *stubSwitcher) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	registerHandlers(mux, nil, sw, slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func postSwitch(t *testing.T, srv *httptest.Server, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := srv.Client().Post(srv.URL+"/switch", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("post switch: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestPostSwitch_accepted(t *testing.T) {
	sw := &stubSwitcher{statusErr: context.DeadlineExceeded}
	resp := postSwitch(t, serveHandlers(t, sw), switchBody{Profile: "alice"})
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected 202, got %d", resp.StatusCode)
	}
	if sw.lastReq != "alice" {
		t.Errorf("expected alice, got %q", sw.lastReq)
	}
}

func TestPostSwitch_rejectsWhenUsersOnline(t *testing.T) {
	sw := &stubSwitcher{status: foundrystatus.Status{Active: true, Users: 2}}
	resp := postSwitch(t, serveHandlers(t, sw), switchBody{Profile: "alice"})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if sw.lastReq != "" {
		t.Errorf("switch should not have been requested, got %q", sw.lastReq)
	}
}

func TestPostSwitch_forceBypassesGuard(t *testing.T) {
	sw := &stubSwitcher{status: foundrystatus.Status{Active: true, Users: 2}}
	resp := postSwitch(t, serveHandlers(t, sw), switchBody{Profile: "alice", Force: true})
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202 with force, got %d", resp.StatusCode)
	}
	if sw.lastReq != "alice" {
		t.Errorf("expected alice, got %q", sw.lastReq)
	}
}

func TestGetStatus_enrichedWhenOnline(t *testing.T) {
	sw := &stubSwitcher{active: "alice", version: "14.0.0", status: foundrystatus.Status{
		Active: true, Version: "13.351", World: "my-world", System: "projectfu",
		SystemVersion: "4.16.1", Users: 3, UptimeMS: 6230770,
	}}
	srv := serveHandlers(t, sw)
	resp, err := srv.Client().Get(srv.URL + "/status")
	if err != nil {
		t.Fatalf("get status: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var got statusResponse
	if err = json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !got.Online || got.World != "my-world" || got.Users != 3 || got.Version != "13.351" {
		t.Errorf("expected enriched live status, got %+v", got)
	}
}

func TestGetStatus_offlineFallsBackToConfigured(t *testing.T) {
	sw := &stubSwitcher{active: "alice", version: "14.0.0", statusErr: context.DeadlineExceeded}
	srv := serveHandlers(t, sw)
	resp, err := srv.Client().Get(srv.URL + "/status")
	if err != nil {
		t.Fatalf("get status: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var got statusResponse
	if err = json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Online || got.Version != "14.0.0" {
		t.Errorf("expected offline with configured version, got %+v", got)
	}
}
