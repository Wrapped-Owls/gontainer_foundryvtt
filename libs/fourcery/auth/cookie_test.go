package auth

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"golang.org/x/net/publicsuffix"
)

func makeTestJar(t *testing.T) http.CookieJar {
	t.Helper()
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		t.Fatal(err)
	}
	return jar
}

func TestHasSessionCookieFalse(t *testing.T) {
	jar := makeTestJar(t)
	if hasSessionCookie(jar) {
		t.Error("empty jar should not have session cookie")
	}
}

func TestHasSessionCookieTrue(t *testing.T) {
	jar := makeTestJar(t)
	u, _ := url.Parse(BaseURL)
	jar.SetCookies(u, []*http.Cookie{{Name: "sessionid", Value: "abc"}})
	if !hasSessionCookie(jar) {
		t.Error("jar with sessionid should return true")
	}
}

func TestExportImportCookiesRoundTrip(t *testing.T) {
	jar := makeTestJar(t)
	u, _ := url.Parse(BaseURL)
	jar.SetCookies(u, []*http.Cookie{
		{Name: "csrftoken", Value: "csrf-val", Path: "/"},
		{Name: "sessionid", Value: "sess-val", Path: "/"},
	})
	saved := exportCookies(jar)
	if len(saved) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(saved))
	}

	jar2 := makeTestJar(t)
	importCookies(jar2, saved)
	for _, c := range jar2.Cookies(u) {
		switch c.Name {
		case "csrftoken":
			if c.Value != "csrf-val" {
				t.Errorf("csrftoken value = %q", c.Value)
			}
		case "sessionid":
			if c.Value != "sess-val" {
				t.Errorf("sessionid value = %q", c.Value)
			}
		}
	}
}
