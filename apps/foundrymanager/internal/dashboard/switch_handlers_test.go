package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/foundrystatus"
)

func TestPostSwitch_accepted(t *testing.T) {
	t.Parallel()

	sup := &stubSupervisor{statusErr: context.DeadlineExceeded}
	resp := postSwitch(t, serveHandlers(t, sup, nil, nil), switchBody{Profile: profAlice})
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected 202, got %d", resp.StatusCode)
	}
	if sup.lastReq != profAlice {
		t.Errorf("expected alice, got %q", sup.lastReq)
	}
}

func TestPostSwitch_rejectsWhenUsersOnline(t *testing.T) {
	t.Parallel()

	sup := &stubSupervisor{status: foundrystatus.Status{Active: true, Users: 2}}
	resp := postSwitch(t, serveHandlers(t, sup, nil, nil), switchBody{Profile: profAlice})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if sup.lastReq != "" {
		t.Errorf("switch should not have been requested, got %q", sup.lastReq)
	}
}

func TestPostSwitch_forceBypassesGuard(t *testing.T) {
	t.Parallel()

	sup := &stubSupervisor{status: foundrystatus.Status{Active: true, Users: 2}}
	resp := postSwitch(
		t,
		serveHandlers(t, sup, nil, nil),
		switchBody{Profile: profAlice, Force: true},
	)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202 with force, got %d", resp.StatusCode)
	}
	if sup.lastReq != profAlice {
		t.Errorf("expected alice, got %q", sup.lastReq)
	}
}

func TestGetStatus_enrichedWhenOnline(t *testing.T) {
	t.Parallel()

	sup := &stubSupervisor{active: profAlice, version: verProfile, status: foundrystatus.Status{
		Active: true, Version: "13.351", World: "my-world", System: "projectfu",
		SystemVersion: "4.16.1", Users: 3, UptimeMS: 6230770,
	}}
	srv := serveHandlers(t, sup, nil, nil)
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
	t.Parallel()

	sup := &stubSupervisor{
		active:    profAlice,
		version:   verProfile,
		statusErr: context.DeadlineExceeded,
	}
	srv := serveHandlers(t, sup, nil, nil)
	resp, err := srv.Client().Get(srv.URL + "/status")
	if err != nil {
		t.Fatalf("get status: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var got statusResponse
	if err = json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Online || got.Version != verProfile {
		t.Errorf("expected offline with configured version, got %+v", got)
	}
}

func TestPostRestart(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		supervisor   stubSupervisor
		body         string
		wantStatus   int
		wantRestarts int
	}{
		{
			name:         "restarts when nobody is connected",
			supervisor:   stubSupervisor{statusErr: context.DeadlineExceeded},
			body:         forceFalse,
			wantStatus:   http.StatusAccepted,
			wantRestarts: 1,
		},
		{
			name:       "refuses while players are online",
			supervisor: stubSupervisor{status: foundrystatus.Status{Active: true, Users: 3}},
			body:       forceFalse,
			wantStatus: http.StatusConflict,
		},
		{
			name:         "force restarts anyway",
			supervisor:   stubSupervisor{status: foundrystatus.Status{Active: true, Users: 3}},
			body:         `{"force":true}`,
			wantStatus:   http.StatusAccepted,
			wantRestarts: 1,
		},
		{
			name:         "an empty body means force:false",
			supervisor:   stubSupervisor{statusErr: context.DeadlineExceeded},
			body:         "",
			wantStatus:   http.StatusAccepted,
			wantRestarts: 1,
		},
		{
			name:       "a malformed body is rejected",
			supervisor: stubSupervisor{statusErr: context.DeadlineExceeded},
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "a restart with no running session is refused, not accepted",
			supervisor: stubSupervisor{
				statusErr:  context.DeadlineExceeded,
				restartErr: errors.New("procloop: no running session"),
			},
			body:         forceFalse,
			wantStatus:   http.StatusConflict,
			wantRestarts: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sup := testCase.supervisor
			srv := serveHandlers(t, &sup, nil, nil)
			resp, err := srv.Client().Post(
				srv.URL+"/restart", "application/json", strings.NewReader(testCase.body),
			)
			if err != nil {
				t.Fatalf("post restart: %v", err)
			}
			t.Cleanup(func() { _ = resp.Body.Close() })

			if resp.StatusCode != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, testCase.wantStatus)
			}
			if sup.restartCalls != testCase.wantRestarts {
				t.Fatalf("restart calls = %d, want %d", sup.restartCalls, testCase.wantRestarts)
			}
		})
	}
}
