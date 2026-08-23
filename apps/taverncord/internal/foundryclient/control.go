package foundryclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

// Switch with force false gets a 409 when users are connected.
func (c *Client) Switch(ctx context.Context, name string, force bool) error {
	callCtx, cancel := withDashboardTimeout(ctx)
	defer cancel()

	body := switchBody{Profile: name, Force: force}
	_, err := jsonhttp.Request[struct{}, switchBody](callCtx, c.cfg, jsonhttp.RequestConfig[switchBody]{
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
	callCtx, cancel := withDashboardTimeout(ctx)
	defer cancel()

	body := restartBody{Force: force}
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

// Download calls POST /versions/download to acquire a Foundry version. It returns
// the dashboard error message verbatim when the download cannot be satisfied.
// It is bounded by downloadRequestTimeout, not the short one the others use.
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
