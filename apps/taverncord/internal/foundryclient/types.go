package foundryclient

import (
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

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

type versionsResp struct {
	Active    string   `json:"active"`
	Installed []string `json:"installed"`
}

type downloadBody struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

type profileDetailResp struct {
	Name         string `json:"name"`
	Label        string `json:"label"`
	DataPath     string `json:"dataPath"`
	Version      string `json:"version"`
	World        string `json:"world"`
	ManifestPath string `json:"manifestPath"`
	HasAdminKey  bool   `json:"hasAdminKey"`
}

type logsResp struct {
	Lines []string `json:"lines"`
}

type eventItemResp struct {
	Time    time.Time `json:"time"`
	Kind    string    `json:"kind"`
	Message string    `json:"message"`
}

type eventsResp struct {
	Events []eventItemResp `json:"events"`
	Next   int             `json:"next"`
}

type restartBody struct {
	Force bool `json:"force"`
}

type profileBody struct {
	Name    string `json:"name,omitempty"`
	Label   string `json:"label,omitempty"`
	Version string `json:"version,omitempty"`
	World   string `json:"world,omitempty"`
}

type errorResp struct {
	Error string `json:"error"`
}
