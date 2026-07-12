package foundrystatus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch_decodesStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/status" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"active":true,"version":"13.351","world":"my-world",` +
			`"system":"projectfu","systemVersion":"4.16.1","users":3,"uptime":6230770}`))
	}))
	defer srv.Close()

	got, err := NewClient(srv.Client()).Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Active || got.World != "my-world" || got.Users != 3 || got.UptimeMS != 6230770 {
		t.Errorf("unexpected status: %+v", got)
	}
	if got.System != "projectfu" || got.SystemVersion != "4.16.1" || got.Version != "13.351" {
		t.Errorf("unexpected status: %+v", got)
	}
}

func TestFetch_serverError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	if _, err := NewClient(srv.Client()).Fetch(context.Background(), srv.URL); err == nil {
		t.Fatal("expected error on 500 response")
	}
}
