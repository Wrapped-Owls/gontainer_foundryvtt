package foundryclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

func (c *Client) Switch(ctx context.Context, name string, interrupt command.Interrupt) error {
	callCtx, cancel := withDashboardTimeout(ctx)
	defer cancel()

	body := switchBody{Profile: name, Force: interrupt == command.InterruptAlways}
	_, err := jsonhttp.Request[struct{}, switchBody](
		callCtx,
		c.cfg,
		jsonhttp.RequestConfig[switchBody]{
			Method: http.MethodPost,
			Path:   "/switch",
			Body:   &body,
			OnStatus: map[int]func(*http.Response) error{
				http.StatusBadRequest: decodeError,
				http.StatusConflict:   decodeError,
				http.StatusAccepted:   func(_ *http.Response) error { return nil },
			},
		},
	)
	return err
}

func (c *Client) Restart(ctx context.Context, interrupt command.Interrupt) error {
	callCtx, cancel := withDashboardTimeout(ctx)
	defer cancel()

	body := restartBody{Force: interrupt == command.InterruptAlways}
	_, err := jsonhttp.Request[struct{}, restartBody](
		callCtx,
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

func (c *Client) Download(ctx context.Context, version, url string) error {
	body := downloadBody{Version: version, URL: url}
	_, err := jsonhttp.Request[struct{}, downloadBody](
		ctx, // uses downloadRequestTimeout, not withDashboardTimeout
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
