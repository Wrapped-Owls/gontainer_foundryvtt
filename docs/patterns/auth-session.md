# `fourcery` auth session

How to authenticate with the FoundryVTT API and reuse the session across calls.

## Overview

`libs/fourcery/auth` provides a cookie-based `Session` backed by an `*http.Client` with
a cookie jar. A `Session` is obtained once via `auth.Login` and then reused for all subsequent
API calls.

## Login flow

```go
sess, err := auth.Login(ctx, auth.Credentials{
    Username: cfg.Install.Username,
    Password: cfg.Install.Password,
})
if err != nil {
    return fmt.Errorf("foundry login: %w", err)
}
```

`auth.Login` POST-encodes the credentials to the FoundryVTT website. On success, the session
cookies are stored in the `Session`'s cookie jar.

## Using the session

Pass `sess.Client()` as the `HTTP` field of `jsonhttp.ClientConfig`:

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
    ...
)
```

The authenticated cookies are sent automatically on each request.

## Session lifetime

A `Session` is valid for the duration of a single `activate.Prepare` run. Do not store it
across invocations. The `sessionSource` in `libs/fourcery/source/session.go` wraps the login
logic and is invoked only when `FOUNDRY_USERNAME`/`FOUNDRY_PASSWORD` or `FOUNDRY_SESSION` are set.

## Skipping auth (custom release URL)

When `FOUNDRY_RELEASE_URL` is set and a local artefact already matches the desired version,
the controller reuses it without any network call. When a URL download is required it fetches
the presigned URL directly — no `auth.Login` is called.

## Testing

Inject a fake `auth.Session` using the `session.NewWith` constructor (or a mock
`HTTPDoer` — see [`jsonhttp.md`](jsonhttp.md)):

```go
fake := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
    // return a canned response
})}
sess := auth.NewSession(fake, "test-agent")
```

## See also

- [`../rules/http-clients.md`](../rules/http-clients.md) — `fourcery` layout.
- [`jsonhttp.md`](jsonhttp.md) — typed `jsonhttp.Request` call pattern.
- [`../rules/security.md`](../rules/security.md) — credential handling.
