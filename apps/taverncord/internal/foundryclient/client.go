package foundryclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

// Client calls the foundrymanager dashboard REST API.
// It implements command.FoundryClient.
type Client struct {
	cfg jsonhttp.ClientConfig
}

// New creates a Client targeting the given base URL (e.g. "http://foundryvtt:30002").
func New(baseURL string) *Client {
	return &Client{cfg: jsonhttp.ClientConfig{
		BaseURL: baseURL,
		HTTP:    &http.Client{},
	}}
}

// ListProfiles calls GET /profiles and returns the profile list with the active profile name.
func (c *Client) ListProfiles(ctx context.Context) (command.ProfilesData, error) {
	resp, err := jsonhttp.Request[profilesResp, struct{}](
		ctx,
		c.cfg,
		jsonhttp.RequestConfig[struct{}]{
			Method: http.MethodGet,
			Path:   "/profiles",
		},
	)
	if err != nil {
		return command.ProfilesData{}, err
	}
	return command.ProfilesData{Active: resp.Active, Profiles: resp.Profiles}, nil
}

// Switch calls POST /switch to request a profile change.
// Returns an error on non-202 responses, including the dashboard error message when available.
func (c *Client) Switch(ctx context.Context, name string) error {
	body := switchBody{Profile: name}
	_, err := jsonhttp.Request[struct{}, switchBody](ctx, c.cfg, jsonhttp.RequestConfig[switchBody]{
		Method: http.MethodPost,
		Path:   "/switch",
		Body:   &body,
		OnStatus: map[int]func(*http.Response) error{
			http.StatusBadRequest: func(r *http.Response) error {
				var e errorResp
				if jsonErr := json.NewDecoder(r.Body).Decode(&e); jsonErr == nil && e.Error != "" {
					return fmt.Errorf("%s", e.Error)
				}
				return fmt.Errorf("bad request")
			},
			http.StatusAccepted: func(_ *http.Response) error { return nil },
		},
	})
	return err
}

// Status calls GET /status and returns the active profile name and Foundry version.
func (c *Client) Status(ctx context.Context) (command.StatusData, error) {
	resp, err := jsonhttp.Request[statusResp, struct{}](
		ctx,
		c.cfg,
		jsonhttp.RequestConfig[struct{}]{
			Method: http.MethodGet,
			Path:   "/status",
		},
	)
	if err != nil {
		return command.StatusData{}, err
	}
	return command.StatusData{Active: resp.Active, Version: resp.Version}, nil
}
