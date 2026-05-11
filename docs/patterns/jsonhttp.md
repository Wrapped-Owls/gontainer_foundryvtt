# `jsonhttp` typed HTTP calls

How to add a new outbound HTTP API call using `libs/foundrykit/jsonhttp`.

## When to use

Use `jsonhttp.Request` whenever code in this repo makes an HTTP call and needs to decode a JSON
response. This is the sole HTTP client helper — do not use `http.Get` or hand-roll
`json.NewDecoder(resp.Body).Decode(...)`.

## Defining a typed response

Add a named struct for the response in the sub-package that owns the API call:

```go
// libs/fourcery/release/types.go
type releaseURLResp struct {
    URL string `json:"url"`
}
```

## Making the call

```go
result, err := jsonhttp.Request[releaseURLResp, struct{}](ctx,
    jsonhttp.ClientConfig{
        BaseURL: auth.BaseURL,
        Headers: map[string]string{
            "User-Agent": sess.UserAgent,
            "Referer":    auth.BaseURL,
        },
        HTTP: sess.Client(),
    },
    jsonhttp.RequestConfig[struct{}]{
        Method: http.MethodGet,
        Path:   "/releases/download?build=361&platform=node&response_type=json",
        OnStatus: map[int]func(*http.Response) error{
            http.StatusUnauthorized: func(_ *http.Response) error {
                return ErrUnauthorized
            },
        },
    },
)
if err != nil {
    return "", err
}
```

## Sending a body

Use the `Body` field of `RequestConfig` with a concrete struct type:

```go
type loginRequest struct {
    UserID   string `json:"userId"`
    Password string `json:"password"`
}

result, err := jsonhttp.Request[loginResponse, loginRequest](ctx, cc,
    jsonhttp.RequestConfig[loginRequest]{
        Method: http.MethodPost,
        Path:   "/auth/local",
        Body:   &loginRequest{UserID: username, Password: password},
    },
)
```

## Testing

Inject a fake `HTTPDoer` into `ClientConfig.HTTP` that returns a prepared `*http.Response`:

```go
type fakeDoer struct{ resp *http.Response }

func (f fakeDoer) Do(_ *http.Request) (*http.Response, error) { return f.resp, nil }
```

Build the response body with `io.NopCloser(bytes.NewReader([]byte(`{"url":"https://..."}`)))`.

## See also

- [`../rules/transport.md`](../rules/transport.md) — typed HTTP rule.
- [`../rules/http-clients.md`](../rules/http-clients.md) — `fourcery` client overview.
