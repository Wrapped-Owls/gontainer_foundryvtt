// Package foundryclient implements command.FoundryClient via HTTP calls to the
// foundrymanager dashboard API.
package foundryclient

import "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"

type profilesResp struct {
	Active   string            `json:"active"`
	Profiles []profile.Profile `json:"profiles"`
}

type statusResp struct {
	Active  string `json:"active"`
	Version string `json:"version"`
}

type switchBody struct {
	Profile string `json:"profile"`
}

type errorResp struct {
	Error string `json:"error"`
}
