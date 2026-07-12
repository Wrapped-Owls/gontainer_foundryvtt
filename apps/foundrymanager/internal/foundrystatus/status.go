package foundrystatus

import (
	"context"
	"net/http"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

type Status struct {
	Active        bool
	Version       string
	World         string
	System        string
	SystemVersion string
	Users         int
	UptimeMS      int64
}

type apiStatus struct {
	Active        bool   `json:"active"`
	Version       string `json:"version"`
	World         string `json:"world"`
	System        string `json:"system"`
	SystemVersion string `json:"systemVersion"`
	Users         int    `json:"users"`
	Uptime        int64  `json:"uptime"`
}

type Client struct {
	http jsonhttp.HTTPDoer
}

func NewClient(doer jsonhttp.HTTPDoer) *Client {
	return &Client{http: doer}
}

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
