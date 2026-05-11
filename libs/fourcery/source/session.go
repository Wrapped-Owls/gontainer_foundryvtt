package source

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/archive"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/auth"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/release"
)

// sessionSource is "I know the version and have credentials → fetch a
// presigned URL via auth + release and then download it." It is the
// authenticated counterpart to urlSource.
type sessionSource struct {
	version  string
	session  string
	username string
	password string
}

// NewSession constructs a sessionSource. Either session or
// username+password must be non-empty; version must be non-empty.
func NewSession(version, session, username, password string) Source {
	return &sessionSource{
		version:  version,
		session:  session,
		username: username,
		password: password,
	}
}

func (s *sessionSource) Kind() Kind { return KindSession }

func (s *sessionSource) Describe() string { return "authenticated session" }

func (s *sessionSource) Probe(_ context.Context) (string, error) {
	if s.version == "" {
		return "", ErrVersionUnknown
	}
	return s.version, nil
}

func (s *sessionSource) Materialise(ctx context.Context, dst string) (Result, error) {
	if s.version == "" {
		return Result{}, fmt.Errorf("%w: version", ErrEmptyInput)
	}
	sess, err := s.login(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("session login: %w", err)
	}
	url, err := release.Fetch(ctx, sess, s.version, release.FetchOptions{})
	if err != nil {
		return Result{}, fmt.Errorf("session release fetch: %w", err)
	}
	zipPath, _, err := downloadToTemp(ctx, sess.Client(), url)
	if err != nil {
		return Result{}, fmt.Errorf("session download: %w", err)
	}
	defer func() { _ = os.Remove(zipPath) }()
	if _, err = archive.Extract(zipPath, dst); err != nil {
		return Result{}, fmt.Errorf("session extract: %w", err)
	}
	return Result{Kind: KindSession, Version: s.version}, nil
}

func (s *sessionSource) login(ctx context.Context) (*auth.Session, error) {
	opts := auth.Options{UserAgent: auth.DefaultUserAgent}
	if s.session != "" {
		sess, err := auth.LoadSession(s.session, opts)
		if err == nil {
			return sess, nil
		}
		if s.username == "" || s.password == "" {
			return nil, err
		}
	}
	if s.username == "" || s.password == "" {
		return nil, errors.New("session: no credentials")
	}
	return auth.Login(ctx, s.username, s.password, opts)
}
