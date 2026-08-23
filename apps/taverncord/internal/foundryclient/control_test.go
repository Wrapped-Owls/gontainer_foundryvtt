package foundryclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

func TestSwitch_accepted(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

func postRestart(
	t *testing.T,
	status int,
	body errorResp,
	interrupt command.Interrupt,
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

	return gotPath, sent.Force, New(srv.URL).Restart(context.Background(), interrupt)
}

func TestRestartAccepted(t *testing.T) {
	t.Parallel()

	for _, interrupt := range []command.Interrupt{command.InterruptWhenIdle, command.InterruptAlways} {
		path, sentForce, err := postRestart(t, http.StatusAccepted, errorResp{}, interrupt)
		if err != nil {
			t.Fatalf("interrupt=%q: unexpected error: %v", interrupt, err)
		}
		if path != "/restart" {
			t.Fatalf("posted to %q, want /restart", path)
		}
		if wantForce := interrupt == command.InterruptAlways; sentForce != wantForce {
			t.Fatalf("force sent as %v, want %v", sentForce, wantForce)
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

			_, _, err := postRestart(t, testCase.status, testCase.body, command.InterruptWhenIdle)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), testCase.wantMessage) {
				t.Fatalf("err = %q, want it to mention %q", err.Error(), testCase.wantMessage)
			}
		})
	}
}
