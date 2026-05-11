package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const csrfPage = `<html><body><form><input type="hidden" name="csrfmiddlewaretoken" value="ABC123"></form></body></html>`

const welcomePage = `<html><body><div id="login-welcome"><a href="/community/Atropos/">Hi</a></div></body></html>`

func newFakeServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			http.SetCookie(w, &http.Cookie{Name: "csrftoken", Value: "csrf-cookie", Path: "/"})
			_, _ = w.Write([]byte(csrfPage))
		case http.MethodPost:
			_ = r.ParseForm()
			if r.Form.Get("csrfmiddlewaretoken") != "ABC123" {
				http.Error(w, "bad csrf", 400)
				return
			}
			if r.Form.Get("password") != "right" {
				_, _ = w.Write([]byte("<html><body>Try again</body></html>"))
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "sessionid", Value: "logged-in", Path: "/"})
			_, _ = w.Write([]byte(welcomePage))
		}
	})
	return httptest.NewServer(mux)
}

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func roundTripperRedirecting(target string) http.RoundTripper {
	return rtFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(strings.TrimPrefix(target, "http://"), "https://")
		return http.DefaultTransport.RoundTrip(req)
	})
}

func TestLoginSuccess(t *testing.T) {
	srv := newFakeServer(t)
	defer srv.Close()

	sess, err := Login(context.Background(), "Atropos", "right", Options{
		HTTPClient: &http.Client{Transport: roundTripperRedirecting(srv.URL)},
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if sess.Username != "atropos" {
		t.Errorf("username = %q, want atropos (lowercased)", sess.Username)
	}
}

func TestLoginBadCredentials(t *testing.T) {
	srv := newFakeServer(t)
	defer srv.Close()
	_, err := Login(context.Background(), "Atropos", "wrong", Options{
		HTTPClient: &http.Client{Transport: roundTripperRedirecting(srv.URL)},
	})
	if err == nil {
		t.Fatal("expected error on bad credentials")
	}
}

func TestFindCSRFToken(t *testing.T) {
	tok, err := findCSRFToken([]byte(csrfPage))
	if err != nil || tok != "ABC123" {
		t.Fatalf("got %q err=%v", tok, err)
	}
	if _, err := findCSRFToken([]byte("<html></html>")); err == nil {
		t.Fatal("expected error when token missing")
	}
}

func TestExtractUsername(t *testing.T) {
	u, err := extractUsername([]byte(welcomePage))
	if err != nil || u != "atropos" {
		t.Fatalf("got %q err=%v", u, err)
	}
	if _, err := extractUsername([]byte("<html></html>")); err == nil {
		t.Fatal("expected error when link missing")
	}
}
