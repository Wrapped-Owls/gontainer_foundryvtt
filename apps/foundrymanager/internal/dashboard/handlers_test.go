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
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

const (
	profAlice    = "alice"
	profBob      = "bob"
	verFoundry13 = "13.351.0"
	verFoundry14 = "14.361.0"
	verProfile   = "14.0.0"
	forceFalse   = `{"force":false}`
)

type stubSupervisor struct {
	active       string
	version      string
	lastReq      string
	err          error
	status       foundrystatus.Status
	statusErr    error
	restartCalls int
	restartErr   error
}

func (s *stubSupervisor) Active() string  { return s.active }
func (s *stubSupervisor) Version() string { return s.version }
func (s *stubSupervisor) RequestSwitch(name string) error {
	s.lastReq = name
	return s.err
}

func (s *stubSupervisor) RequestRestart() error {
	s.restartCalls++
	return s.restartErr
}

func (s *stubSupervisor) FoundryStatus(_ context.Context) (foundrystatus.Status, error) {
	return s.status, s.statusErr
}

type stubVersions struct {
	installed   []string
	installErr  error
	downloadErr error
	lastVersion string
	lastURL     string
}

func (v *stubVersions) Installed(_ context.Context) ([]string, error) {
	return v.installed, v.installErr
}

func (v *stubVersions) Download(_ context.Context, version, url string) error {
	v.lastVersion, v.lastURL = version, url
	return v.downloadErr
}

type stubProfiles struct {
	profiles       []profile.Profile
	getProfile     profile.Profile
	getOK          bool
	createErr      error
	updateErr      error
	deleteErr      error
	lastCreate     profile.Profile
	lastUpdateName string
	lastDelete     string
}

func (s *stubProfiles) ListProfiles() []profile.Profile           { return s.profiles }
func (s *stubProfiles) GetProfile(string) (profile.Profile, bool) { return s.getProfile, s.getOK }

func (s *stubProfiles) CreateProfile(
	p profile.Profile,
) error {
	s.lastCreate = p
	if s.createErr != nil {
		return s.createErr
	}
	p.AdminKey = ""
	p.AdminPasswordSalt = ""
	p.ManifestPath = ""
	s.getProfile = p
	s.getOK = true
	return nil
}

func (s *stubProfiles) DeleteProfile(
	name string,
) error {
	s.lastDelete = name
	return s.deleteErr
}

func (s *stubProfiles) UpdateProfile(name string, _ profile.Profile) error {
	s.lastUpdateName = name
	return s.updateErr
}

func serveHandlers(
	t *testing.T,
	sup *stubSupervisor,
	vm *stubVersions,
	ps *stubProfiles,
) *httptest.Server {
	t.Helper()
	if vm == nil {
		vm = &stubVersions{}
	}
	if ps == nil {
		ps = &stubProfiles{}
	}
	mux := http.NewServeMux()
	registerHandlers(mux, sup, vm, ps, slog.New(slog.DiscardHandler))
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

func TestOnlinePlayers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		supervisor stubSupervisor
		want       int
	}{
		{
			name:       "an unreachable server counts nobody",
			supervisor: stubSupervisor{statusErr: context.DeadlineExceeded},
		},
		{
			name:       "a server with no world active counts nobody",
			supervisor: stubSupervisor{status: foundrystatus.Status{Active: false, Users: 4}},
		},
		{
			name:       "connected players are counted",
			supervisor: stubSupervisor{status: foundrystatus.Status{Active: true, Users: 4}},
			want:       4,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sup := testCase.supervisor
			if got := onlinePlayers(context.Background(), &sup); got != testCase.want {
				t.Fatalf("onlinePlayers() = %d, want %d", got, testCase.want)
			}
		})
	}
}
