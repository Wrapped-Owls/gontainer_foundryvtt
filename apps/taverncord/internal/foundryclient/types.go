// Package foundryclient implements command.FoundryClient via HTTP calls to the
// foundrymanager dashboard API.
package foundryclient

import "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"

type profilesResp struct {
	Active   string            `json:"active"`
	Profiles []profile.Profile `json:"profiles"`
}

type statusResp struct {
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

type switchBody struct {
	Profile string `json:"profile"`
	Force   bool   `json:"force"`
}

type errorResp struct {
	Error string `json:"error"`
}
