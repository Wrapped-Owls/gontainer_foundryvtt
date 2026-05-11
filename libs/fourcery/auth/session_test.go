package auth

import (
	"context"
	"net/http"
	"testing"
)

func TestSessionRoundTrip(t *testing.T) {
	srv := newFakeServer(t)
	defer srv.Close()
	sess, err := Login(context.Background(), "Atropos", "right", Options{
		HTTPClient: &http.Client{Transport: roundTripperRedirecting(srv.URL)},
	})
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/sess.json"
	if err := sess.Save(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadSession(path, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Username != sess.Username {
		t.Errorf("username mismatch")
	}
	if !hasSessionCookie(loaded.Jar()) {
		t.Errorf("sessionid cookie did not survive round-trip")
	}
}
