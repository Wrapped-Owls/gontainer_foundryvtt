// Package health implements the Foundry liveness probe used by the Docker
// HEALTHCHECK directive.
package health

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultPort    = 30000
	DefaultTimeout = 5 * time.Second
)

const (
	envSSLCert     = "FOUNDRY_SSL_CERT"
	envSSLKey      = "FOUNDRY_SSL_KEY"
	envRoutePrefix = "FOUNDRY_ROUTE_PREFIX"
)

// Probe describes the URL to hit. Build with FromEnv() in production.
type Probe struct {
	URL     string
	Timeout time.Duration
	// Insecure mirrors curl --insecure: skip TLS cert verification.
	Insecure bool
}

func Default() Probe {
	return Probe{
		URL:      fmt.Sprintf("http://localhost:%d/api/status", DefaultPort),
		Timeout:  DefaultTimeout,
		Insecure: false,
	}
}

// FromEnv mirrors check_health.sh's URL construction:
//
//   - protocol = "https" iff both FOUNDRY_SSL_CERT and FOUNDRY_SSL_KEY are set
//   - path     = /<FOUNDRY_ROUTE_PREFIX>/api/status when prefix is set
func FromEnv(env func(string) string) Probe {
	if env == nil {
		env = func(k string) string { return "" }
	}
	proto := "http"
	if env(envSSLCert) != "" && env(envSSLKey) != "" {
		proto = "https"
	}
	prefix := strings.Trim(env(envRoutePrefix), "/")
	path := "/api/status"
	if prefix != "" {
		path = "/" + prefix + path
	}
	return Probe{
		URL:      fmt.Sprintf("%s://localhost:%d%s", proto, DefaultPort, path),
		Timeout:  DefaultTimeout,
		Insecure: proto == "https",
	}
}

// Check executes the probe. Returns nil iff the server replied 2xx.
func Check(ctx context.Context, p Probe) error {
	if p.Timeout <= 0 {
		p.Timeout = DefaultTimeout
	}
	cctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodGet, p.URL, nil)
	if err != nil {
		return err
	}
	tr := &http.Transport{}
	if p.Insecure {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		} //nolint:gosec // matches curl --insecure
	}
	client := &http.Client{Transport: tr, Timeout: p.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("health: %s returned %s", p.URL, resp.Status)
	}
	return nil
}
