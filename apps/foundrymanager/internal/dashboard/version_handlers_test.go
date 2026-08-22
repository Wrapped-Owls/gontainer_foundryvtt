package dashboard

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func postJSON(t *testing.T, srv *httptest.Server, path string, body any) *http.Response {
	t.Helper()

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("encode %s body: %v", path, err)
	}
	resp, err := srv.Client().Post(srv.URL+path, "application/json", bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("post %s: %v", path, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestGetVersionsListsInstalled(t *testing.T) {
	t.Parallel()

	srv := serveHandlers(
		t,
		&stubSupervisor{version: verFoundry14},
		&stubVersions{installed: []string{verFoundry14, verFoundry13}},
		nil,
	)
	resp, err := srv.Client().Get(srv.URL + "/versions")
	if err != nil {
		t.Fatalf("get versions: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var got versionsResponse
	if err = json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Active != verFoundry14 || len(got.Installed) != 2 {
		t.Fatalf("unexpected versions response: %+v", got)
	}
}

func TestPostDownload(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		versions    stubVersions
		body        downloadBody
		wantStatus  int
		wantVersion string
		wantMessage string
	}{
		{
			name:        "a download is accepted and its arguments forwarded",
			body:        downloadBody{Version: verFoundry14, URL: "https://signed"},
			wantStatus:  http.StatusAccepted,
			wantVersion: verFoundry14,
		},
		{
			name:       "a body with no version is refused",
			body:       downloadBody{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:        "a failed download relays the reason",
			versions:    stubVersions{downloadErr: errors.New("no source for 9.9.9")},
			body:        downloadBody{Version: "9.9.9"},
			wantStatus:  http.StatusBadGateway,
			wantVersion: "9.9.9",
			wantMessage: "no source",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			versions := testCase.versions
			srv := serveHandlers(t, &stubSupervisor{}, &versions, nil)
			resp := postJSON(t, srv, "/versions/download", testCase.body)

			if resp.StatusCode != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, testCase.wantStatus)
			}
			if versions.lastVersion != testCase.wantVersion {
				t.Fatalf(
					"forwarded version %q, want %q",
					versions.lastVersion,
					testCase.wantVersion,
				)
			}
			if testCase.wantMessage == "" {
				return
			}
			var got errorResponse
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !strings.Contains(got.Error, testCase.wantMessage) {
				t.Fatalf("error = %q, want it to mention %q", got.Error, testCase.wantMessage)
			}
		})
	}
}
