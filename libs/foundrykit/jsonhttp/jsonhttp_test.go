package jsonhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testResp struct {
	Name string `json:"name"`
}

type testBody struct {
	Value string `json:"value"`
}

func testConfig(handler http.Handler) (ClientConfig, *httptest.Server) {
	srv := httptest.NewServer(handler)
	cc := ClientConfig{
		BaseURL: srv.URL,
		HTTP:    srv.Client(),
	}
	return cc, srv
}

func TestRequest_DecodesResponse(t *testing.T) {
	cc, srv := testConfig(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testResp{Name: "ok"})
	}))
	defer srv.Close()

	resp, err := Request[testResp, any](context.Background(), cc, RequestConfig[any]{
		Method: http.MethodGet,
		Path:   "/test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != "ok" {
		t.Fatalf("name = %q, want %q", resp.Name, "ok")
	}
}

func TestRequest_EncodesBody(t *testing.T) {
	var gotBody testBody
	cc, srv := testConfig(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("expected Content-Type: application/json")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testResp{Name: "ok"})
	}))
	defer srv.Close()

	body := testBody{Value: "hello"}
	_, err := Request[testResp](context.Background(), cc, RequestConfig[testBody]{
		Method: http.MethodPost,
		Path:   "/post",
		Body:   &body,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody.Value != "hello" {
		t.Fatalf("value = %q, want %q", gotBody.Value, "hello")
	}
}

func TestRequest_DefaultErrorOnHTTPStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"400 bad request", http.StatusBadRequest},
		{"404 not found", http.StatusNotFound},
		{"500 server error", http.StatusInternalServerError},
		{"502 bad gateway", http.StatusBadGateway},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc, srv := testConfig(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			_, err := Request[testResp, any](context.Background(), cc, RequestConfig[any]{
				Method: http.MethodGet,
				Path:   "/test",
			})
			if err == nil {
				t.Fatalf("expected error for status %d", tt.status)
			}
		})
	}
}

func TestRequest_OnStatusCallback(t *testing.T) {
	sentinel := errors.New("custom not found")

	tests := []struct {
		name     string
		callback func(*http.Response) error
		wantErr  error
	}{
		{
			name:     "returns custom error",
			callback: func(_ *http.Response) error { return sentinel },
			wantErr:  sentinel,
		},
		{
			name:     "returns nil suppresses error",
			callback: func(_ *http.Response) error { return nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc, srv := testConfig(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))
			defer srv.Close()

			_, err := Request[testResp, any](context.Background(), cc, RequestConfig[any]{
				Method: http.MethodGet,
				Path:   "/item",
				OnStatus: map[int]func(*http.Response) error{
					404: tt.callback,
				},
			})
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRequest_ConnectionError(t *testing.T) {
	// Use a closed server to simulate connection refused.
	cc, srv := testConfig(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	srv.Close()

	_, err := Request[testResp, any](context.Background(), cc, RequestConfig[any]{
		Method: http.MethodGet,
		Path:   "/status",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRequest_HeadersAttached(t *testing.T) {
	var gotHeader string
	cc, srv := testConfig(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Custom")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testResp{})
	}))
	defer srv.Close()

	cc.Headers = map[string]string{"X-Custom": "value"}

	_, _ = Request[testResp, any](context.Background(), cc, RequestConfig[any]{
		Method: http.MethodGet,
		Path:   "/test",
	})
	if gotHeader != "value" {
		t.Fatalf("header = %q, want %q", gotHeader, "value")
	}
}
