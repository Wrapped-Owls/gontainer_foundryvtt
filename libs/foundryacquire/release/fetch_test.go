package release

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryacquire/auth"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func makeSession(t *testing.T, srv *httptest.Server) *auth.Session {
	t.Helper()
	jar, _ := cookiejar.New(nil)
	transport := rtFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(srv.URL, "http://")
		return http.DefaultTransport.RoundTrip(req)
	})
	client := &http.Client{Transport: transport, Jar: jar}
	tmp := t.TempDir() + "/s.json"
	stub := &auth.Session{Username: "atropos", UserAgent: "test"}
	if err := stub.Save(tmp); err != nil {
		t.Fatal(err)
	}
	loaded, err := auth.LoadSession(tmp, auth.Options{HTTPClient: client, UserAgent: "test"})
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}

func TestFetchSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases/download" || r.URL.Query().Get("build") != "361" {
			t.Errorf("unexpected request %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"url":"https://s3/foundry-14.361.zip","lifetime":300}`))
	}))
	defer srv.Close()
	sess := makeSession(t, srv)

	url, err := Fetch(context.Background(), sess, "14.361", FetchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://s3/foundry-14.361.zip" {
		t.Errorf("url = %q", url)
	}
}

func TestFetchEmptyURLErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"url":"","lifetime":0}`))
	}))
	defer srv.Close()
	sess := makeSession(t, srv)

	if _, err := Fetch(context.Background(), sess, "14.361", FetchOptions{}); !errors.Is(
		err,
		ErrEmptyURL,
	) {
		t.Fatalf("expected ErrEmptyURL, got %v", err)
	}
}

func TestFetchRetries(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			http.Error(w, "boom", http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(`{"url":"https://ok"}`))
	}))
	defer srv.Close()
	sess := makeSession(t, srv)

	noSleep := func(ctx context.Context, _ time.Duration) error { return nil }
	url, err := Fetch(context.Background(), sess, "14.361",
		FetchOptions{Retries: 5, Sleep: noSleep})
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://ok" || atomic.LoadInt32(&calls) != 3 {
		t.Errorf("url=%q calls=%d", url, calls)
	}
}

func TestBuildNumber(t *testing.T) {
	cases := map[string]string{"14.361": "361", "0.7.4": "4", "999": "999"}
	for in, want := range cases {
		got, err := buildNumber(in)
		if err != nil || got != want {
			t.Errorf("buildNumber(%q) = %q err=%v", in, got, err)
		}
	}
	if _, err := buildNumber(""); !errors.Is(err, ErrNoBuildNumber) {
		t.Errorf("expected ErrNoBuildNumber for empty input")
	}
	if _, err := buildNumber("14."); !errors.Is(err, ErrNoBuildNumber) {
		t.Errorf("expected ErrNoBuildNumber for trailing dot")
	}
}
