package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
