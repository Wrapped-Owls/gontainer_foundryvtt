package jsonhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPDoer abstracts HTTP client calls. *http.Client satisfies this interface.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientConfig wraps the base URL, default headers, and HTTP transport.
type ClientConfig struct {
	BaseURL string
	Headers map[string]string
	HTTP    HTTPDoer
}

// RequestConfig describes a single HTTP request with an optional typed body
// and per-status callbacks.
type RequestConfig[Body any] struct {
	Method   string
	Path     string
	Body     *Body
	OnStatus map[int]func(*http.Response) error
}

// Request sends an HTTP request and decodes the JSON response into Resp.
//
// Flow:
//  1. Encode Body as JSON if non-nil
//  2. Build request with ClientConfig.BaseURL + RequestConfig.Path
//  3. Attach default headers from ClientConfig
//  4. If OnStatus has a callback for the response status code, call it
//  5. If no callback matches and status/100 is 4 or 5, return an error
//  6. Otherwise decode response body into *Resp
func Request[Resp any, Body any](
	ctx context.Context,
	cc ClientConfig,
	rc RequestConfig[Body],
) (*Resp, error) {
	var body io.Reader
	if rc.Body != nil {
		b, err := json.Marshal(rc.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, rc.Method, cc.BaseURL+rc.Path, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range cc.Headers {
		req.Header.Set(k, v)
	}

	resp, err := cc.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if rc.OnStatus != nil {
		if fn, ok := rc.OnStatus[resp.StatusCode]; ok {
			return nil, fn(resp)
		}
	}

	switch resp.StatusCode / 100 {
	case 4, 5:
		return nil, fmt.Errorf("%s %s: status %d", rc.Method, rc.Path, resp.StatusCode)
	}

	var result Resp
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}
