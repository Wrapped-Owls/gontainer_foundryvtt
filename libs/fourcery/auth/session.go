package auth

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"golang.org/x/net/publicsuffix"
)

type Session struct {
	Username  string        `json:"username"`
	UserAgent string        `json:"user_agent"`
	Cookies   []SavedCookie `json:"cookies"`

	jar    http.CookieJar
	client *http.Client
}

type SavedCookie struct {
	Name    string    `json:"name"`
	Value   string    `json:"value"`
	Domain  string    `json:"domain"`
	Path    string    `json:"path"`
	Expires time.Time `json:"expires"`
}

func (s *Session) Client() *http.Client { return s.client }

func (s *Session) Jar() http.CookieJar { return s.jar }

func (s *Session) Save(path string) error {
	const secretPerm fs.FileMode = 0o600
	if s.jar != nil {
		s.Cookies = exportCookies(s.jar)
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, secretPerm)
}

func LoadSession(path string, opts Options) (*Session, error) {
	if opts.UserAgent == "" {
		opts.UserAgent = DefaultUserAgent
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Session
	if err = json.Unmarshal(b, &s); err != nil {
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
	const defaultTimeout = 30 * time.Second
	if base == nil {
		return &http.Client{Jar: jar, Timeout: defaultTimeout}
	}
	cp := *base
	cp.Jar = jar
	return &cp
}
