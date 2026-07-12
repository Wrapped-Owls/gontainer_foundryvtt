package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/foundrystatus"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
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

// stubVersions implements dashboard.VersionManager for testing.
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

// stubProfiles implements dashboard.ProfileStore for testing.
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
func (s *stubProfiles) CreateProfile(p profile.Profile) error     { s.lastCreate = p; return s.createErr }
func (s *stubProfiles) DeleteProfile(name string) error           { s.lastDelete = name; return s.deleteErr }

func (s *stubProfiles) UpdateProfile(name string, _ profile.Profile) error {
	s.lastUpdateName = name
	return s.updateErr
}

func serveHandlers(t *testing.T, sw *stubSwitcher, vm *stubVersions, ps *stubProfiles) *httptest.Server {
	t.Helper()
	if vm == nil {
		vm = &stubVersions{}
	}
	if ps == nil {
		ps = &stubProfiles{}
	}
	mux := http.NewServeMux()
	registerHandlers(mux, sw, vm, ps, slog.New(slog.DiscardHandler))
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
	resp := postSwitch(t, serveHandlers(t, sw, nil, nil), switchBody{Profile: "alice"})
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected 202, got %d", resp.StatusCode)
	}
	if sw.lastReq != "alice" {
		t.Errorf("expected alice, got %q", sw.lastReq)
	}
}

func TestPostSwitch_rejectsWhenUsersOnline(t *testing.T) {
	sw := &stubSwitcher{status: foundrystatus.Status{Active: true, Users: 2}}
	resp := postSwitch(t, serveHandlers(t, sw, nil, nil), switchBody{Profile: "alice"})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if sw.lastReq != "" {
		t.Errorf("switch should not have been requested, got %q", sw.lastReq)
	}
}

func TestPostSwitch_forceBypassesGuard(t *testing.T) {
	sw := &stubSwitcher{status: foundrystatus.Status{Active: true, Users: 2}}
	resp := postSwitch(t, serveHandlers(t, sw, nil, nil), switchBody{Profile: "alice", Force: true})
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
	srv := serveHandlers(t, sw, nil, nil)
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

func TestGetVersions_listsInstalled(t *testing.T) {
	sw := &stubSwitcher{version: "14.361.0"}
	vm := &stubVersions{installed: []string{"14.361.0", "13.351.0"}}
	srv := serveHandlers(t, sw, vm, nil)
	resp, err := srv.Client().Get(srv.URL + "/versions")
	if err != nil {
		t.Fatalf("get versions: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var got versionsResponse
	if err = json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Active != "14.361.0" || len(got.Installed) != 2 {
		t.Errorf("unexpected versions response: %+v", got)
	}
}

func TestPostDownload_accepted(t *testing.T) {
	vm := &stubVersions{}
	srv := serveHandlers(t, &stubSwitcher{}, vm, nil)
	b, _ := json.Marshal(downloadBody{Version: "14.361.0", URL: "https://signed"})
	resp, err := srv.Client().Post(srv.URL+"/versions/download", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("post download: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
	if vm.lastVersion != "14.361.0" || vm.lastURL != "https://signed" {
		t.Errorf("download args not forwarded: %+v", vm)
	}
}

func TestPostDownload_missingVersion(t *testing.T) {
	srv := serveHandlers(t, &stubSwitcher{}, &stubVersions{}, nil)
	b, _ := json.Marshal(downloadBody{})
	resp, err := srv.Client().Post(srv.URL+"/versions/download", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("post download: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestPostDownload_failureRelaysError(t *testing.T) {
	vm := &stubVersions{downloadErr: errors.New("no source for 9.9.9")}
	srv := serveHandlers(t, &stubSwitcher{}, vm, nil)
	b, _ := json.Marshal(downloadBody{Version: "9.9.9"})
	resp, err := srv.Client().Post(srv.URL+"/versions/download", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("post download: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", resp.StatusCode)
	}
	var got errorResponse
	if err = json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.Contains(got.Error, "no source") {
		t.Errorf("expected relayed error, got %q", got.Error)
	}
}

func TestGetStatus_offlineFallsBackToConfigured(t *testing.T) {
	sw := &stubSwitcher{active: "alice", version: "14.0.0", statusErr: context.DeadlineExceeded}
	srv := serveHandlers(t, sw, nil, nil)
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

func TestGetProfile_redactsSecrets(t *testing.T) {
	ps := &stubProfiles{getOK: true, getProfile: profile.Profile{
		Name: "alice", DataPath: "/d", AdminKey: "s3cret", World: "w",
	}}
	srv := serveHandlers(t, &stubSwitcher{}, nil, ps)
	resp, err := srv.Client().Get(srv.URL + "/profiles/alice")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), "s3cret") {
		t.Errorf("admin key leaked in response: %s", body)
	}
	var got profileDetail
	if err = json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.World != "w" || !got.HasAdminKey {
		t.Errorf("unexpected detail: %+v", got)
	}
}

func TestGetProfile_notFound(t *testing.T) {
	srv := serveHandlers(t, &stubSwitcher{}, nil, &stubProfiles{getOK: false})
	resp, err := srv.Client().Get(srv.URL + "/profiles/ghost")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestPostProfile_created(t *testing.T) {
	ps := &stubProfiles{}
	srv := serveHandlers(t, &stubSwitcher{}, nil, ps)
	b, _ := json.Marshal(profile.Profile{Name: "bob", DataPath: "/d/bob"})
	resp, err := srv.Client().Post(srv.URL+"/profiles", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("post profile: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if ps.lastCreate.Name != "bob" {
		t.Errorf("create not forwarded: %+v", ps.lastCreate)
	}
}

func TestPostProfile_conflict(t *testing.T) {
	ps := &stubProfiles{createErr: profile.ErrExists}
	srv := serveHandlers(t, &stubSwitcher{}, nil, ps)
	b, _ := json.Marshal(profile.Profile{Name: "bob", DataPath: "/d"})
	resp, err := srv.Client().Post(srv.URL+"/profiles", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("post profile: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409, got %d", resp.StatusCode)
	}
}

func TestPutProfile_notFound(t *testing.T) {
	ps := &stubProfiles{updateErr: profile.ErrNotFound}
	srv := serveHandlers(t, &stubSwitcher{}, nil, ps)
	b, _ := json.Marshal(profile.Profile{World: "new"})
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/profiles/ghost", bytes.NewReader(b))
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("put profile: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeleteProfile_activeRefused(t *testing.T) {
	ps := &stubProfiles{deleteErr: profile.ErrInvalid}
	srv := serveHandlers(t, &stubSwitcher{}, nil, ps)
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/profiles/alice", nil)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("delete profile: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestDeleteProfile_ok(t *testing.T) {
	ps := &stubProfiles{}
	srv := serveHandlers(t, &stubSwitcher{}, nil, ps)
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/profiles/bob", nil)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("delete profile: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if ps.lastDelete != "bob" {
		t.Errorf("delete not forwarded: %q", ps.lastDelete)
	}
}
