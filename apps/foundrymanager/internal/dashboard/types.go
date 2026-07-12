package dashboard

import (
	"context"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/foundrystatus"
)

// Switcher is the minimal interface consumed by dashboard HTTP handlers.
// Manager satisfies this interface implicitly — no import of manager/ needed.
type Switcher interface {
	RequestSwitch(name string) error
	Active() string
	Version() string
	FoundryStatus(ctx context.Context) (foundrystatus.Status, error)
}

type profileRef struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type profilesResponse struct {
	Active   string       `json:"active"`
	Profiles []profileRef `json:"profiles"`
}

type switchBody struct {
	Profile string `json:"profile"`
	Force   bool   `json:"force"`
}

type statusResponse struct {
	Active        string `json:"active"`
	Version       string `json:"version"`
	Online        bool   `json:"online"`
	WorldActive   bool   `json:"worldActive"`
	World         string `json:"world"`
	System        string `json:"system"`
	SystemVersion string `json:"systemVersion"`
	Users         int    `json:"users"`
	UptimeMS      int64  `json:"uptimeMs"`
}

type errorResponse struct {
	Error string `json:"error"`
}
