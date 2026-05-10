package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromEnvDefaults(t *testing.T) {
	p := FromEnv(func(string) string { return "" })
	if p.URL != "http://localhost:30000/api/status" {
		t.Errorf("default URL: %q", p.URL)
	}
	if p.Insecure {
		t.Error("expected insecure=false for plain http")
	}
}

func TestFromEnvHTTPSAndPrefix(t *testing.T) {
	env := map[string]string{
		"FOUNDRY_SSL_CERT":     "/x.crt",
		"FOUNDRY_SSL_KEY":      "/x.key",
		"FOUNDRY_ROUTE_PREFIX": "/foundry/",
	}
	p := FromEnv(func(k string) string { return env[k] })
	if p.URL != "https://localhost:30000/foundry/api/status" {
		t.Errorf("URL: %q", p.URL)
	}
	if !p.Insecure {
		t.Error("expected insecure=true for https")
	}
}

func TestCheckSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/status" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	if err := Check(context.Background(), Probe{URL: srv.URL + "/api/status"}); err != nil {
		t.Fatal(err)
	}
}

func TestCheckFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	if err := Check(context.Background(), Probe{URL: srv.URL + "/api/status"}); err == nil {
		t.Fatal("expected error on 500")
	}
}
