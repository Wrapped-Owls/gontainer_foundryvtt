package foundryclient

import (
	"context"
	"net/http"
	"net/url"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

// GetProfile calls GET /profiles/{name} and returns the non-secret profile view.
func (c *Client) GetProfile(ctx context.Context, name string) (command.ProfileInfo, error) {
	resp, err := jsonhttp.Request[profileDetailResp, struct{}](
		ctx,
		c.cfg,
		jsonhttp.RequestConfig[struct{}]{
			Method: http.MethodGet,
			Path:   profilePath(name),
			OnStatus: map[int]func(*http.Response) error{
				http.StatusNotFound: decodeError,
			},
		},
	)
	if err != nil {
		return command.ProfileInfo{}, err
	}
	return command.ProfileInfo{
		Name:         resp.Name,
		Label:        resp.Label,
		DataPath:     resp.DataPath,
		Version:      resp.Version,
		World:        resp.World,
		ManifestPath: resp.ManifestPath,
		HasAdminKey:  resp.HasAdminKey,
	}, nil
}

// CreateProfile calls POST /profiles to create a new profile.
func (c *Client) CreateProfile(ctx context.Context, p command.ProfileInput) error {
	body := toBody(p)
	_, err := jsonhttp.Request[struct{}, profileBody](
		ctx,
		c.cfg,
		jsonhttp.RequestConfig[profileBody]{
			Method: http.MethodPost,
			Path:   profilesPath,
			Body:   &body,
			OnStatus: map[int]func(*http.Response) error{
				http.StatusCreated:    func(_ *http.Response) error { return nil },
				http.StatusBadRequest: decodeError,
				http.StatusConflict:   decodeError,
			},
		},
	)
	return err
}

// UpdateProfile calls PUT /profiles/{name} to merge non-empty fields.
func (c *Client) UpdateProfile(ctx context.Context, name string, p command.ProfileInput) error {
	body := toBody(p)
	_, err := jsonhttp.Request[struct{}, profileBody](
		ctx,
		c.cfg,
		jsonhttp.RequestConfig[profileBody]{
			Method: http.MethodPut,
			Path:   profilePath(name),
			Body:   &body,
			OnStatus: map[int]func(*http.Response) error{
				http.StatusOK:         func(_ *http.Response) error { return nil },
				http.StatusBadRequest: decodeError,
				http.StatusNotFound:   decodeError,
			},
		},
	)
	return err
}

// DeleteProfile calls DELETE /profiles/{name}.
func (c *Client) DeleteProfile(ctx context.Context, name string) error {
	_, err := jsonhttp.Request[struct{}, struct{}](ctx, c.cfg, jsonhttp.RequestConfig[struct{}]{
		Method: http.MethodDelete,
		Path:   profilePath(name),
		OnStatus: map[int]func(*http.Response) error{
			http.StatusNoContent:  func(_ *http.Response) error { return nil },
			http.StatusBadRequest: decodeError,
			http.StatusNotFound:   decodeError,
		},
	})
	return err
}

func profilePath(name string) string {
	return profilesPath + "/" + url.PathEscape(name)
}

func toBody(p command.ProfileInput) profileBody {
	return profileBody{
		Name:         p.Name,
		Label:        p.Label,
		DataPath:     p.DataPath,
		Version:      p.Version,
		World:        p.World,
		ManifestPath: p.ManifestPath,
	}
}
