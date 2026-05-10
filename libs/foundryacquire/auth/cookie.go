package auth

import (
	"net/http"
	"net/url"
)

func hasSessionCookie(jar http.CookieJar) bool {
	u, _ := url.Parse(BaseURL)
	for _, c := range jar.Cookies(u) {
		if c.Name == "sessionid" && c.Value != "" {
			return true
		}
	}
	return false
}

func exportCookies(jar http.CookieJar) []SavedCookie {
	u, _ := url.Parse(BaseURL)
	cs := jar.Cookies(u)
	out := make([]SavedCookie, len(cs))
	for i, c := range cs {
		out[i] = SavedCookie{
			Name:   c.Name,
			Value:  c.Value,
			Domain: u.Hostname(),
			Path:   "/",
		}
	}
	return out
}

func importCookies(jar http.CookieJar, in []SavedCookie) {
	u, _ := url.Parse(BaseURL)
	cs := make([]*http.Cookie, len(in))
	for i, sc := range in {
		cs[i] = &http.Cookie{
			Name:    sc.Name,
			Value:   sc.Value,
			Domain:  sc.Domain,
			Path:    sc.Path,
			Expires: sc.Expires,
		}
	}
	jar.SetCookies(u, cs)
}
