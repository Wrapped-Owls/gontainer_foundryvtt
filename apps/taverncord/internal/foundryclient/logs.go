package foundryclient

import (
	"context"
	"net/http"
	"strconv"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

// Logs calls GET /logs and returns the most recent captured log lines.
func (c *Client) Logs(ctx context.Context, tail int) (command.LogsData, error) {
	resp, err := jsonhttp.Request[logsResp, struct{}](
		ctx,
		c.cfg,
		jsonhttp.RequestConfig[struct{}]{
			Method: http.MethodGet,
			Path:   "/logs?tail=" + strconv.Itoa(tail),
		},
	)
	if err != nil {
		return command.LogsData{}, err
	}
	return command.LogsData{Lines: resp.Lines}, nil
}

// Events calls GET /events and returns detected events at or after since.
func (c *Client) Events(ctx context.Context, since int) (command.EventsData, error) {
	resp, err := jsonhttp.Request[eventsResp, struct{}](
		ctx,
		c.cfg,
		jsonhttp.RequestConfig[struct{}]{
			Method: http.MethodGet,
			Path:   "/events?since=" + strconv.Itoa(since),
		},
	)
	if err != nil {
		return command.EventsData{}, err
	}
	events := make([]command.EventItem, len(resp.Events))
	for i, e := range resp.Events {
		events[i] = command.EventItem{Time: e.Time, Kind: e.Kind, Message: e.Message}
	}
	return command.EventsData{Events: events, Next: resp.Next}, nil
}
