package foundryclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

func TestListProfiles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/profiles" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(profilesResp{
			Active: "alice",
			Profiles: []profile.Profile{
				{Name: "alice", Label: "Alice"},
				{Name: "bob", Label: "Bob"},
			},
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	profiles, err := c.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profiles.Active != "alice" {
		t.Errorf("expected active=alice, got %q", profiles.Active)
	}
	if len(profiles.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles.Profiles))
	}
}

func TestSwitch_accepted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	if err := New(
		srv.URL,
	).Switch(context.Background(), "bob", command.InterruptWhenIdle); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSwitch_conflictOnlineUsers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(errorResp{Error: "2 user(s) currently online"})
	}))
	defer srv.Close()

	err := New(srv.URL).Switch(context.Background(), "bob", command.InterruptWhenIdle)
	if err == nil {
		t.Fatal("expected error when users are online")
	}
	if !strings.Contains(err.Error(), "online") {
		t.Errorf("expected online message, got %q", err.Error())
	}
}

func TestSwitch_badRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResp{Error: "unknown profile"})
	}))
	defer srv.Close()

	err := New(srv.URL).Switch(context.Background(), "nobody", command.InterruptWhenIdle)
	if err == nil {
		t.Fatal("expected error for bad request")
	}
}

func TestVersions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/versions" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(versionsResp{
			Active: "14.361.0", Installed: []string{"14.361.0", "13.351.0"},
		})
	}))
	defer srv.Close()

	versions, err := New(srv.URL).Versions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if versions.Active != "14.361.0" || len(versions.Installed) != 2 {
		t.Errorf("unexpected data: %+v", versions)
	}
}

func TestDownload_accepted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	if err := New(srv.URL).Download(context.Background(), "14.361.0", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDownload_badGatewayRelaysError(t *testing.T) {
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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/logs" || r.URL.Query().Get("tail") != "10" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(logsResp{Lines: []string{"a", "b"}})
	}))
	defer srv.Close()

	logs, err := New(srv.URL).Logs(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs.Lines) != 2 {
		t.Errorf("expected 2 lines, got %+v", logs.Lines)
	}
}

func TestEvents(t *testing.T) {
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

	events, err := New(srv.URL).Events(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if events.Next != 6 || len(events.Events) != 1 || events.Events[0].Kind != "crash" {
		t.Errorf("unexpected events data: %+v", events)
	}
}

func TestStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(statusResp{
			Active:        "alice",
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

	status, err := New(srv.URL).Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Active != "alice" || status.Version != "13.351" {
		t.Errorf("unexpected data: %+v", status)
	}
	if !status.Online || status.World != "my-world" || status.Users != 2 ||
		status.SystemVersion != "4.16.1" {
		t.Errorf("expected live status fields, got %+v", status)
	}
}
