package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
)

// Login performs the CSRF dance + form post. On success the returned
// Session is ready to use.
func Login(ctx context.Context, username, password string, opts Options) (*Session, error) {
	if opts.UserAgent == "" {
		opts.UserAgent = DefaultUserAgent
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	client := buildClient(opts.HTTPClient, jar)

	csrf, err := fetchCSRF(ctx, client, opts.UserAgent)
	if err != nil {
		return nil, err
	}

	form := url.Values{
		"csrfmiddlewaretoken": {csrf},
		"next":                {"/"},
		"username":            {strings.ToLower(username)},
		"password":            {password},
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		BaseURL+LoginPath,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}
	setCommonHeaders(req, opts.UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", BaseURL+LoginPath)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth: POST login: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("auth: login HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if !hasSessionCookie(jar) {
		return nil, ErrSessionCookieMissing
	}
	canonical, err := extractUsername(body)
	if err != nil {
		return nil, err
	}

	return &Session{
		Username:  canonical,
		UserAgent: opts.UserAgent,
		jar:       jar,
		client:    client,
	}, nil
}

func setCommonHeaders(req *http.Request, ua string) {
	req.Header.Set("User-Agent", ua)
	req.Header.Set("DNT", "1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Referer", BaseURL)
}

func fetchCSRF(ctx context.Context, client *http.Client, ua string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, BaseURL+LoginPath, nil)
	if err != nil {
		return "", err
	}
	setCommonHeaders(req, ua)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth: GET CSRF page: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("auth: CSRF page HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return findCSRFToken(body)
}

func findCSRFToken(body []byte) (string, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	var token string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if token != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "input" {
			var name, value string
			for _, a := range n.Attr {
				switch a.Key {
				case "name":
					name = a.Val
				case "value":
					value = a.Val
				}
			}
			if name == "csrfmiddlewaretoken" {
				token = strings.TrimSpace(value)
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if token == "" {
		return "", ErrCSRFNotFound
	}
	return token, nil
}

func extractUsername(body []byte) (string, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	var found string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if found != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" && strings.HasPrefix(a.Val, "/community/") {
					rest := a.Val[len("/community/"):]
					if i := strings.IndexAny(rest, "/?#"); i >= 0 {
						rest = rest[:i]
					}
					if rest != "" {
						found = strings.ToLower(rest)
						return
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if found == "" {
		return "", ErrCommunityLinkMissing
	}
	return found, nil
}
