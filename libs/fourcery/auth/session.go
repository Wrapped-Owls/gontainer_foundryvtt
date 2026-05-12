package auth

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
)

// Session is the result of a successful Login or LoadSession call. It
// carries the authenticated cookie jar and the canonical (lowercase)
// username derived from the user's community profile link.
type Session struct {
	Username  string        `json:"username"`
	UserAgent string        `json:"user_agent"`
	Cookies   []SavedCookie `json:"cookies"`

	jar    http.CookieJar
	client *http.Client
}

// SavedCookie is the on-disk shape used by Session.Save / LoadSession.
type SavedCookie struct {
	Name    string    `json:"name"`
	Value   string    `json:"value"`
	Domain  string    `json:"domain"`
	Path    string    `json:"path"`
	Expires time.Time `json:"expires"`
}

// Client returns the authenticated *http.Client (cookies attached). Other
// packages (release, license) accept *Session and call this internally.
func (s *Session) Client() *http.Client { return s.client }

// Jar exposes the underlying cookie jar (mostly useful for tests).
func (s *Session) Jar() http.CookieJar { return s.jar }

// Save serialises the session to JSON. Use this to share authentication
// across the build/runtime boundary.
func (s *Session) Save(path string) error {
	if s.jar != nil {
		s.Cookies = exportCookies(s.jar)
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, fsperm.Secret)
}

// LoadSession restores a saved session. The HTTP client is rebuilt so
// the caller can pass HTTPClient.Transport overrides via opts.
func LoadSession(path string, opts Options) (*Session, error) {
	if opts.UserAgent == "" {
		opts.UserAgent = DefaultUserAgent
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	importCookies(jar, s.Cookies)
	s.jar = jar
	s.client = buildClient(opts.HTTPClient, jar)
	if opts.UserAgent != "" {
		s.UserAgent = opts.UserAgent
	}
	return &s, nil
}

func buildClient(base *http.Client, jar http.CookieJar) *http.Client {
	if base == nil {
		return &http.Client{Jar: jar, Timeout: 30 * time.Second}
	}
	cp := *base
	cp.Jar = jar
	return &cp
}
