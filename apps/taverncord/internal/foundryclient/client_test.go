package foundryclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

const (
	profAlice    = "alice"
	verFoundry14 = "14.361.0"
)

func TestListProfiles(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/profiles" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(profilesResp{
			Active: profAlice,
			Profiles: []profile.Profile{
				{Name: profAlice, Label: "Alice"},
				{Name: "bob", Label: "Bob"},
			},
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	data, err := c.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Active != profAlice {
		t.Errorf("expected active=alice, got %q", data.Active)
	}
	if len(data.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(data.Profiles))
	}
}

func TestSwitch_accepted(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	if err := New(srv.URL).Switch(context.Background(), "bob", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSwitch_conflictOnlineUsers(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(errorResp{Error: "2 user(s) currently online"})
	}))
	defer srv.Close()

	err := New(srv.URL).Switch(context.Background(), "bob", false)
	if err == nil {
		t.Fatal("expected error when users are online")
	}
	if !strings.Contains(err.Error(), "online") {
		t.Errorf("expected online message, got %q", err.Error())
	}
}

func TestSwitch_badRequest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResp{Error: "unknown profile"})
	}))
	defer srv.Close()

	err := New(srv.URL).Switch(context.Background(), "nobody", false)
	if err == nil {
		t.Fatal("expected error for bad request")
	}
}

func TestVersions(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/versions" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(versionsResp{
			Active: verFoundry14, Installed: []string{verFoundry14, "13.351.0"},
		})
	}))
	defer srv.Close()

	data, err := New(srv.URL).Versions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Active != verFoundry14 || len(data.Installed) != 2 {
		t.Errorf("unexpected data: %+v", data)
	}
}

func TestDownload_accepted(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	if err := New(srv.URL).Download(context.Background(), verFoundry14, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDownload_badGatewayRelaysError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(errorResp{Error: "no source for 9.9.9"})
	}))
	defer srv.Close()

	err := New(srv.URL).Download(context.Background(), "9.9.9", "")
	if err == nil || !strings.Contains(err.Error(), "no source") {
		t.Errorf("expected relayed error, got %v", err)
	}
}

func TestLogs(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/logs" || r.URL.Query().Get("tail") != "10" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(logsResp{Lines: []string{"a", "b"}})
	}))
	defer srv.Close()

	data, err := New(srv.URL).Logs(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.Lines) != 2 {
		t.Errorf("expected 2 lines, got %+v", data.Lines)
	}
}

func TestEvents(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("since") != "5" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(eventsResp{
			Events: []eventItemResp{{Kind: "crash", Message: "boom"}},
			Next:   6,
		})
	}))
	defer srv.Close()

	data, err := New(srv.URL).Events(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Next != 6 || len(data.Events) != 1 || data.Events[0].Kind != "crash" {
		t.Errorf("unexpected events data: %+v", data)
	}
}

func TestStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(statusResp{
			Active:        profAlice,
			Version:       "13.351",
			Online:        true,
			WorldActive:   true,
			World:         "my-world",
			System:        "projectfu",
			SystemVersion: "4.16.1",
			Users:         2,
			UptimeMS:      6230770,
		})
	}))
	defer srv.Close()

	data, err := New(srv.URL).Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Active != profAlice || data.Version != "13.351" {
		t.Errorf("unexpected data: %+v", data)
	}
	if !data.Online || data.World != "my-world" || data.Users != 2 ||
		data.SystemVersion != "4.16.1" {
		t.Errorf("expected live status fields, got %+v", data)
	}
}

func postRestart(
	t *testing.T,
	status int,
	body errorResp,
	force bool,
) (gotPath string, gotForce bool, err error) {
	t.Helper()

	var sent restartBody
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&sent)
		w.WriteHeader(status)
		if body.Error != "" {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
	defer srv.Close()

	return gotPath, sent.Force, New(srv.URL).Restart(context.Background(), force)
}

func TestRestartAccepted(t *testing.T) {
	t.Parallel()

	for _, force := range []bool{false, true} {
		path, sentForce, err := postRestart(t, http.StatusAccepted, errorResp{}, force)
		if err != nil {
			t.Fatalf("force=%v: unexpected error: %v", force, err)
		}
		if path != "/restart" {
			t.Fatalf("posted to %q, want /restart", path)
		}
		if sentForce != force {
			t.Fatalf("force sent as %v, want %v", sentForce, force)
		}
	}
}

func TestRestartRelaysTheManagersReason(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		status      int
		body        errorResp
		wantMessage string
	}{
		{
			name:        "players online",
			status:      http.StatusConflict,
			body:        errorResp{Error: "2 user(s) currently online"},
			wantMessage: "online",
		},
		{
			name:        "a failed reset",
			status:      http.StatusInternalServerError,
			body:        errorResp{Error: "failed to request a restart"},
			wantMessage: "failed to request",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := postRestart(t, testCase.status, testCase.body, false)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), testCase.wantMessage) {
				t.Fatalf("err = %q, want it to mention %q", err.Error(), testCase.wantMessage)
			}
		})
	}
}
