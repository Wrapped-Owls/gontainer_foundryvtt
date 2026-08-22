package foundryclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

type Client struct {
	cfg jsonhttp.ClientConfig
}

var _ command.FoundryClient = (*Client)(nil)

const profilesPath = "/profiles"

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
			Path:   profilesPath,
		},
	)
	if err != nil {
		return command.ProfilesData{}, err
	}
	return command.ProfilesData{Active: resp.Active, Profiles: resp.Profiles}, nil
}

// Switch with force false gets a 409 when users are connected.
func (c *Client) Switch(ctx context.Context, name string, force bool) error {
	body := switchBody{Profile: name, Force: force}
	_, err := jsonhttp.Request[struct{}, switchBody](ctx, c.cfg, jsonhttp.RequestConfig[switchBody]{
		Method: http.MethodPost,
		Path:   "/switch",
		Body:   &body,
		OnStatus: map[int]func(*http.Response) error{
			http.StatusBadRequest: decodeError,
			http.StatusConflict:   decodeError,
			http.StatusAccepted:   func(_ *http.Response) error { return nil },
		},
	})
	return err
}

// Restart with force false gets a 409 when players are connected.
func (c *Client) Restart(ctx context.Context, force bool) error {
	body := restartBody{Force: force}
	_, err := jsonhttp.Request[struct{}, restartBody](
		ctx,
		c.cfg,
		jsonhttp.RequestConfig[restartBody]{
			Method: http.MethodPost,
			Path:   "/restart",
			Body:   &body,
			OnStatus: map[int]func(*http.Response) error{
				http.StatusBadRequest:          decodeError,
				http.StatusConflict:            decodeError,
				http.StatusInternalServerError: decodeError,
				http.StatusAccepted:            func(_ *http.Response) error { return nil },
			},
		},
	)
	return err
}

func decodeError(r *http.Response) error {
	var e errorResp
	if jsonErr := json.NewDecoder(r.Body).Decode(&e); jsonErr == nil && e.Error != "" {
		return fmt.Errorf("%s", e.Error)
	}
	return fmt.Errorf("request rejected with status %d", r.StatusCode)
}

// Versions calls GET /versions and returns the installed versions and the active one.
func (c *Client) Versions(ctx context.Context) (command.VersionsData, error) {
	resp, err := jsonhttp.Request[versionsResp, struct{}](
		ctx,
		c.cfg,
		jsonhttp.RequestConfig[struct{}]{
			Method: http.MethodGet,
			Path:   "/versions",
		},
	)
	if err != nil {
		return command.VersionsData{}, err
	}
	return command.VersionsData{Active: resp.Active, Installed: resp.Installed}, nil
}

// Download calls POST /versions/download to acquire a Foundry version. It returns
// the dashboard error message verbatim when the download cannot be satisfied.
func (c *Client) Download(ctx context.Context, version, url string) error {
	body := downloadBody{Version: version, URL: url}
	_, err := jsonhttp.Request[struct{}, downloadBody](
		ctx,
		c.cfg,
		jsonhttp.RequestConfig[downloadBody]{
			Method: http.MethodPost,
			Path:   "/versions/download",
			Body:   &body,
			OnStatus: map[int]func(*http.Response) error{
				http.StatusBadRequest: decodeError,
				http.StatusBadGateway: decodeError,
				http.StatusAccepted:   func(_ *http.Response) error { return nil },
			},
		},
	)
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
	return command.StatusData{
		Active:        resp.Active,
		Version:       resp.Version,
		Online:        resp.Online,
		WorldActive:   resp.WorldActive,
		World:         resp.World,
		System:        resp.System,
		SystemVersion: resp.SystemVersion,
		Users:         resp.Users,
		UptimeMS:      resp.UptimeMS,
	}, nil
}
