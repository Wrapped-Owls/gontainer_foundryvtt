package foundryclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

func TestListProfiles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/profiles" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(profilesResp{ //nolint:errcheck
			Active: "alice",
			Profiles: []profile.Profile{
				{Name: "alice", Label: "Alice"},
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
	if data.Active != "alice" {
		t.Errorf("expected active=alice, got %q", data.Active)
	}
	if len(data.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(data.Profiles))
	}
}

func TestSwitch_accepted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	if err := New(srv.URL).Switch(context.Background(), "bob"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSwitch_badRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResp{Error: "unknown profile"}) //nolint:errcheck
	}))
	defer srv.Close()

	err := New(srv.URL).Switch(context.Background(), "nobody")
	if err == nil {
		t.Fatal("expected error for bad request")
	}
}

func TestStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(statusResp{Active: "alice", Version: "14.0.0"}) //nolint:errcheck
	}))
	defer srv.Close()

	data, err := New(srv.URL).Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Active != "alice" || data.Version != "14.0.0" {
		t.Errorf("unexpected data: %+v", data)
	}
}
