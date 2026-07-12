package foundrystatus

import (
	"context"
	"net/http"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

// Status is the subset of Foundry's GET /api/status payload the manager consumes.
type Status struct {
	Active        bool
	Version       string
	World         string
	System        string
	SystemVersion string
	Users         int
	UptimeMS      int64
}

// apiStatus is the raw wire shape returned by Foundry's /api/status endpoint.
type apiStatus struct {
	Active        bool   `json:"active"`
	Version       string `json:"version"`
	World         string `json:"world"`
	System        string `json:"system"`
	SystemVersion string `json:"systemVersion"`
	Users         int    `json:"users"`
	Uptime        int64  `json:"uptime"`
}

// Client fetches the live status of a Foundry server over HTTP.
type Client struct {
	http jsonhttp.HTTPDoer
}

// NewClient returns a Client using the given HTTP doer for requests.
func NewClient(doer jsonhttp.HTTPDoer) *Client {
	return &Client{http: doer}
}

// Fetch retrieves the status from the Foundry server reachable at baseURL
// (e.g. "http://127.0.0.1:30000"). The caller is responsible for the request
// deadline via ctx.
func (c *Client) Fetch(ctx context.Context, baseURL string) (Status, error) {
	resp, err := jsonhttp.Request[apiStatus, struct{}](
		ctx,
		jsonhttp.ClientConfig{BaseURL: baseURL, HTTP: c.http},
		jsonhttp.RequestConfig[struct{}]{
			Method: http.MethodGet,
			Path:   "/api/status",
		},
	)
	if err != nil {
		return Status{}, err
	}
	return Status{
		Active:        resp.Active,
		Version:       resp.Version,
		World:         resp.World,
		System:        resp.System,
		SystemVersion: resp.SystemVersion,
		Users:         resp.Users,
		UptimeMS:      resp.Uptime,
	}, nil
}
