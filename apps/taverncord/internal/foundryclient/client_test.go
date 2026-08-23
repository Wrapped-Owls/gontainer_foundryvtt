package foundryclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

const (
	profAlice    = "alice"
	verFoundry14 = "14.361.0"
)

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	c := New("http://example.invalid")

	client, ok := c.cfg.HTTP.(*http.Client)
	if !ok {
		t.Fatalf("HTTP = %T, want *http.Client", c.cfg.HTTP)
	}
	if client.Timeout != downloadRequestTimeout {
		t.Fatalf("HTTP.Timeout = %v, want %v", client.Timeout, downloadRequestTimeout)
	}
}

func TestClientKeepsInjectedHTTPDoer(t *testing.T) {
	t.Parallel()

	injected := &http.Client{Timeout: 5 * time.Second}
	c := &Client{cfg: jsonhttp.ClientConfig{BaseURL: "http://example.invalid", HTTP: injected}}

	if c.cfg.HTTP != jsonhttp.HTTPDoer(injected) {
		t.Fatalf("HTTP = %v, want unchanged injected client", c.cfg.HTTP)
	}
}

type contextCapturingDoer struct {
	gotCtx context.Context
}

func (d *contextCapturingDoer) Do(req *http.Request) (*http.Response, error) {
	d.gotCtx = req.Context()
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
		Header:     make(http.Header),
	}, nil
}

func TestFastCallsUseShortContextTimeout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		call func(c *Client) error
	}{
		{"ListProfiles", func(c *Client) error { _, err := c.ListProfiles(context.Background()); return err }},
		{"Versions", func(c *Client) error { _, err := c.Versions(context.Background()); return err }},
		{"Status", func(c *Client) error { _, err := c.Status(context.Background()); return err }},
		{"GetProfile", func(c *Client) error { _, err := c.GetProfile(context.Background(), profAlice); return err }},
		{"UpdateProfile", func(c *Client) error {
			return c.UpdateProfile(context.Background(), profAlice, command.ProfileInput{})
		}},
		{"Logs", func(c *Client) error { _, err := c.Logs(context.Background(), 1); return err }},
		{"Events", func(c *Client) error { _, err := c.Events(context.Background(), 0); return err }},
		{"Switch", func(c *Client) error { return c.Switch(context.Background(), profAlice, false) }},
		{"Restart", func(c *Client) error { return c.Restart(context.Background(), false) }},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			doer := &contextCapturingDoer{}
			c := &Client{cfg: jsonhttp.ClientConfig{BaseURL: "http://example.invalid", HTTP: doer}}

			before := time.Now()
			if err := testCase.call(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			deadline, hasDeadline := doer.gotCtx.Deadline()
			if !hasDeadline {
				t.Fatal("expected the request context to carry a deadline")
			}
			const slack = 50 * time.Millisecond
			remaining := deadline.Sub(before)
			if remaining <= 0 || remaining > dashboardRequestTimeout+slack {
				t.Fatalf("deadline %v from call start, want within (0, %v]", remaining, dashboardRequestTimeout)
			}
		})
	}
}

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
	profiles, err := c.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profiles.Active != profAlice {
		t.Errorf("expected active=alice, got %q", profiles.Active)
	}
	if len(profiles.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles.Profiles))
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

	versions, err := New(srv.URL).Versions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if versions.Active != verFoundry14 || len(versions.Installed) != 2 {
		t.Errorf("unexpected data: %+v", versions)
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

	logs, err := New(srv.URL).Logs(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs.Lines) != 2 {
		t.Errorf("expected 2 lines, got %+v", logs.Lines)
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

	events, err := New(srv.URL).Events(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if events.Next != 6 || len(events.Events) != 1 || events.Events[0].Kind != "crash" {
		t.Errorf("unexpected events data: %+v", events)
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

	status, err := New(srv.URL).Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Active != profAlice || status.Version != "13.351" {
		t.Errorf("unexpected data: %+v", status)
	}
	if !status.Online || status.World != "my-world" || status.Users != 2 ||
		status.SystemVersion != "4.16.1" {
		t.Errorf("expected live status fields, got %+v", status)
	}
}
