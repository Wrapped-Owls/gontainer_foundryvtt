package procloop

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	fmconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/controller"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
)

func makeRunner(profiles []profile.Profile) *Runner {
	return New(
		State{Profiles: profiles, Version: verProfile},
		"",
		nil,
		nil,
		fmconfig.Config{},
		backoff.Config{},
		slog.Default(),
	)
}

func TestRequestSwitch_unknownProfile(t *testing.T) {
	t.Parallel()

	r := makeRunner([]profile.Profile{{Name: profAlice}})
	if err := r.RequestSwitch("nobody"); err == nil {
		t.Error("expected error for unknown profile")
	}
}

func TestRequestSwitch_knownProfile(t *testing.T) {
	t.Parallel()

	r := makeRunner([]profile.Profile{{Name: profAlice}})
	if err := r.RequestSwitch(profAlice); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestActive_initial(t *testing.T) {
	t.Parallel()

	r := makeRunner(nil)
	if got := r.Active(); got != "" {
		t.Errorf("expected empty initial active, got %q", got)
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()

	r := makeRunner(nil)
	if got := r.Version(); got != verProfile {
		t.Errorf("expected 14.0.0, got %q", got)
	}
}

func TestRequestRestart(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		cacheDir    func(t *testing.T) string
		withSession bool
		wantErr     error
	}{
		{
			name:        "a live session is cancelled with the restart cause",
			cacheDir:    func(t *testing.T) string { return t.TempDir() },
			withSession: true,
		},
		{
			name:        "an absent cache still restarts",
			cacheDir:    func(*testing.T) string { return "" },
			withSession: true,
		},
		{
			name: "a history that cannot be cleared still restarts",
			cacheDir: func(t *testing.T) string {
				dir := t.TempDir()
				blocked := filepath.Join(dir, "backoff_state.json")
				if err := os.MkdirAll(filepath.Join(blocked, "occupied"), 0o755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			withSession: true,
		},
		{
			name:     "no running session is refused instead of silently accepted",
			cacheDir: func(t *testing.T) string { return t.TempDir() },
			wantErr:  ErrNoSession,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			r := &Runner{
				logger:     slog.New(slog.DiscardHandler),
				ctrl:       controller.New(),
				backoffCfg: backoff.Config{CacheDir: testCase.cacheDir(t)},
			}
			var cause error
			if testCase.withSession {
				_, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				r.ctrl.SetCancel(func(err error) { cause = err; cancel(err) })
			}

			err := r.RequestRestart()
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("err = %v, want %v", err, testCase.wantErr)
			}
			if testCase.wantErr != nil {
				return
			}
			if !errors.Is(cause, controller.ErrRestart) {
				t.Fatalf("cancel cause = %v, want %v", cause, controller.ErrRestart)
			}
			select {
			case name := <-r.ctrl.SwitchCh:
				t.Fatalf("a restart must not queue a switch, got %q", name)
			default:
			}
		})
	}
}

func TestRequestRestartDoesNotFireTwiceOnAStaleCancel(t *testing.T) {
	t.Parallel()

	r := &Runner{
		logger:     slog.New(slog.DiscardHandler),
		ctrl:       controller.New(),
		backoffCfg: backoff.Config{},
	}
	_, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	r.ctrl.SetCancel(func(err error) { cancel(err) })

	if err := r.RequestRestart(); err != nil {
		t.Fatalf("first restart: %v", err)
	}
	// The session is gone until runSession registers the next cancel; a second
	// request must not report success against the one that already fired.
	if err := r.RequestRestart(); !errors.Is(err, ErrNoSession) {
		t.Fatalf("second restart err = %v, want %v", err, ErrNoSession)
	}
}
