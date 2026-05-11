# External API clients — `fourcery`

`libs/fourcery` is the typed HTTP client for the FoundryVTT authenticated download API.
It handles cookie-based authentication, release URL resolution, and archive acquisition.

## Layout

```
libs/fourcery/
├── go.mod
├── auth/
│   ├── auth.go          # BaseURL constant, Session type
│   ├── login.go         # Login (username + password → Session)
│   ├── cookie.go        # cookie jar helpers
│   ├── session.go       # Session.Client() — returns an auth'd *http.Client
│   └── *_test.go
└── release/
    ├── types.go          # FetchOptions, releaseURLResp
    ├── fetch.go          # Fetch (Session + version → presigned URL)
    ├── retry.go          # jitter-based retry delay
    └── *_test.go
```

## The `jsonhttp` rule

All HTTP interactions in this library use `libs/foundrykit/jsonhttp.Request[Resp, Body]`:

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
        Path:   path,
    },
)
```

Response types are named structs — `map[string]any` is forbidden for any request or response body.

## Authenticated session

`auth.Login` returns a `*Session` backed by a cookie jar. Pass `sess.Client()` as the `HTTP`
field of `jsonhttp.ClientConfig` to reuse the authenticated cookies on subsequent requests.

## Retry

`release.Fetch` retries with jitter on transient failures up to `FetchOptions.Retries` attempts,
using `backoff.Sleep` from `libs/foundrykit/backoff` to honour context cancellation.

## Adding a new API call

1. Define a named response struct in the relevant sub-package.
2. Wire a `jsonhttp.ClientConfig` using the caller's `*auth.Session`.
3. Call `jsonhttp.Request[YourResp, YourBody]`.
4. Handle specific status codes via `OnStatus` callbacks.

## Forbidden

- `http.Get` / `http.Post` / `http.Client.Do` with manual JSON decode — always use
  `jsonhttp.Request`.
- `map[string]any` for request or response bodies.
- Constructing `*http.Client` directly in call sites — use `sess.Client()`.
- Swallowing retry errors; let them propagate to the activation step.

## See also

- [`transport.md`](transport.md) — typed HTTP client rule.
- [`security.md`](security.md) — credential handling for `FOUNDRY_USERNAME`/`FOUNDRY_PASSWORD`.
