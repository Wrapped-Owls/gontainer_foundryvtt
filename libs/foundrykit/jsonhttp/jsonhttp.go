package jsonhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientConfig struct {
	BaseURL string
	Headers map[string]string
	HTTP    HTTPDoer
}

type RequestConfig[Body any] struct {
	Method   string
	Path     string
	Body     *Body
	OnStatus map[int]func(*http.Response) error
}

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

	const (
		clientErrorClass = 4
		serverErrorClass = 5
	)
	switch resp.StatusCode / 100 {
	case clientErrorClass, serverErrorClass:
		return nil, fmt.Errorf("%s %s: status %d", rc.Method, rc.Path, resp.StatusCode)
	}

	var result Resp
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}
