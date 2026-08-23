package dashboard

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

func TestListProfiles_includesVersionAndWorldNoSecrets(t *testing.T) {
	t.Parallel()

	ps := &stubProfiles{profiles: []profile.Profile{
		{Name: profAlice, Label: "Alice", Version: verProfile, World: "avalon", AdminKey: "s3cret"},
	}}
	srv := serveHandlers(t, &stubSupervisor{active: profAlice}, nil, ps)
	resp, err := srv.Client().Get(srv.URL + "/profiles")
	if err != nil {
		t.Fatalf("list profiles: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), "s3cret") {
		t.Errorf("admin key leaked in list: %s", body)
	}
	var got profilesResponse
	if err = json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got.Profiles) != 1 ||
		got.Profiles[0].Version != verProfile || got.Profiles[0].World != "avalon" {
		t.Errorf("expected version and world in list, got %+v", got.Profiles)
	}
}

func TestGetProfile_redactsSecrets(t *testing.T) {
	t.Parallel()

	ps := &stubProfiles{getOK: true, getProfile: profile.Profile{
		Name: profAlice, DataPath: "/d", AdminKey: "s3cret", World: "w",
	}}
	srv := serveHandlers(t, &stubSupervisor{}, nil, ps)
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
	t.Parallel()

	srv := serveHandlers(t, &stubSupervisor{}, nil, &stubProfiles{getOK: false})
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
	t.Parallel()

	ps := &stubProfiles{}
	srv := serveHandlers(t, &stubSupervisor{}, nil, ps)
	b, _ := json.Marshal(profile.Profile{Name: profBob, DataPath: "/d/bob"})
	resp, err := srv.Client().Post(srv.URL+"/profiles", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("post profile: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if ps.lastCreate.Name != profBob {
		t.Errorf("create not forwarded: %+v", ps.lastCreate)
	}
}

func TestPostProfile_responseReflectsPersistedProfile(t *testing.T) {
	t.Parallel()

	ps := &stubProfiles{}
	srv := serveHandlers(t, &stubSupervisor{}, nil, ps)
	b, _ := json.Marshal(profile.Profile{
		Name: profBob, DataPath: "/d/bob",
		AdminKey: "s3cret", ManifestPath: "/evil/manifest.yaml",
	})
	resp, err := srv.Client().Post(srv.URL+"/profiles", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("post profile: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var got profileDetail
	if err = json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ManifestPath != "" || got.HasAdminKey {
		t.Errorf("response echoes unsanitized input: %+v", got)
	}
}

func TestPostProfile_conflict(t *testing.T) {
	t.Parallel()

	ps := &stubProfiles{createErr: profile.ErrExists}
	srv := serveHandlers(t, &stubSupervisor{}, nil, ps)
	b, _ := json.Marshal(profile.Profile{Name: profBob, DataPath: "/d"})
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
	t.Parallel()

	ps := &stubProfiles{updateErr: profile.ErrNotFound}
	srv := serveHandlers(t, &stubSupervisor{}, nil, ps)
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
	t.Parallel()

	ps := &stubProfiles{deleteErr: profile.ErrInvalid}
	srv := serveHandlers(t, &stubSupervisor{}, nil, ps)
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
	t.Parallel()

	ps := &stubProfiles{}
	srv := serveHandlers(t, &stubSupervisor{}, nil, ps)
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/profiles/bob", nil)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("delete profile: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if ps.lastDelete != profBob {
		t.Errorf("delete not forwarded: %q", ps.lastDelete)
	}
}
