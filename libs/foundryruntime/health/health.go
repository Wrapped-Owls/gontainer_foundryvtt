// Package health implements the Foundry liveness probe used by the Docker
// HEALTHCHECK directive.
package health

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

const (
	DefaultPort    = 30000
	DefaultTimeout = 5 * time.Second
)

// Probe describes the URL to hit.
type Probe struct {
	URL      string
	Timeout  time.Duration
	Insecure bool
}

func Default() Probe {
	return Probe{
		URL:      fmt.Sprintf("http://localhost:%d/api/status", DefaultPort),
		Timeout:  DefaultTimeout,
		Insecure: false,
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
		} //nolint:gosec
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
