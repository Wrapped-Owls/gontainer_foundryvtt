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

func New(baseURL string) *Client {
	return &Client{cfg: jsonhttp.ClientConfig{
		BaseURL: baseURL,
		HTTP:    &http.Client{},
	}}
}

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

func (c *Client) Switch(ctx context.Context, name string, interrupt command.Interrupt) error {
	body := switchBody{Profile: name, Force: interrupt == command.InterruptAlways}
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

func decodeError(r *http.Response) error {
	var e errorResp
	if jsonErr := json.NewDecoder(r.Body).Decode(&e); jsonErr == nil && e.Error != "" {
		return fmt.Errorf("%s", e.Error)
	}
	return fmt.Errorf("request rejected with status %d", r.StatusCode)
}

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
