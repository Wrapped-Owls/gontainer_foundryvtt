package foundryclient

import (
	"context"
	"net/http"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

type Client struct {
	cfg jsonhttp.ClientConfig
}

var _ command.FoundryClient = (*Client)(nil)

const profilesPath = "/profiles"

// dashboardRequestTimeout bounds the fast, synchronous dashboard calls.
const dashboardRequestTimeout = 2 * time.Second

// downloadRequestTimeout bounds Download, whose handler fetches and extracts
// the release before answering. Matches libs/fourcery/source.
const downloadRequestTimeout = 30 * time.Minute

// New creates a Client targeting the given base URL (e.g. "http://foundryvtt:30002").
func New(baseURL string) *Client {
	return &Client{cfg: jsonhttp.ClientConfig{
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: downloadRequestTimeout},
	}}
}

// withDashboardTimeout is for the fast calls; Download needs the long bound.
func withDashboardTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, dashboardRequestTimeout)
}

// ListProfiles calls GET /profiles and returns the profile list with the active profile name.
func (c *Client) ListProfiles(ctx context.Context) (command.ProfilesData, error) {
	callCtx, cancel := withDashboardTimeout(ctx)
	defer cancel()

	resp, err := jsonhttp.Request[profilesResp, struct{}](
		callCtx,
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

// Versions calls GET /versions and returns the installed versions and the active one.
func (c *Client) Versions(ctx context.Context) (command.VersionsData, error) {
	callCtx, cancel := withDashboardTimeout(ctx)
	defer cancel()

	resp, err := jsonhttp.Request[versionsResp, struct{}](
		callCtx,
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

// Status calls GET /status and returns the active profile name and Foundry version.
func (c *Client) Status(ctx context.Context) (command.StatusData, error) {
	callCtx, cancel := withDashboardTimeout(ctx)
	defer cancel()

	resp, err := jsonhttp.Request[statusResp, struct{}](
		callCtx,
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
