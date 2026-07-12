package dashboard

import (
	"context"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/foundrystatus"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/logstore"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

// Switcher is the minimal interface consumed by dashboard HTTP handlers.
// Manager satisfies this interface implicitly — no import of manager/ needed.
type Switcher interface {
	RequestSwitch(name string) error
	Active() string
	Version() string
	FoundryStatus(ctx context.Context) (foundrystatus.Status, error)
}

// VersionManager lists and acquires Foundry installs for the versions endpoints.
type VersionManager interface {
	Installed(ctx context.Context) ([]string, error)
	Download(ctx context.Context, version, url string) error
}

// LogReader exposes recent Foundry log lines and detected events.
type LogReader interface {
	Logs(n int) []string
	Events(cursor int) ([]logstore.Event, int)
}

// ProfileStore reads and mutates the configured profiles for the CRUD endpoints.
type ProfileStore interface {
	ListProfiles() []profile.Profile
	GetProfile(name string) (profile.Profile, bool)
	CreateProfile(p profile.Profile) error
	UpdateProfile(name string, p profile.Profile) error
	DeleteProfile(name string) error
}

type profileRef struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type profilesResponse struct {
	Active   string       `json:"active"`
	Profiles []profileRef `json:"profiles"`
}

// profileDetail is the single-profile response; it deliberately omits the admin
// key and salt so secrets never leave the manager.
type profileDetail struct {
	Name         string `json:"name"`
	Label        string `json:"label"`
	DataPath     string `json:"dataPath"`
	Version      string `json:"version"`
	World        string `json:"world"`
	ManifestPath string `json:"manifestPath"`
	HasAdminKey  bool   `json:"hasAdminKey"`
}

func toDetail(p profile.Profile) profileDetail {
	return profileDetail{
		Name:         p.Name,
		Label:        p.Label,
		DataPath:     p.DataPath,
		Version:      p.Version,
		World:        p.World,
		ManifestPath: p.ManifestPath,
		HasAdminKey:  p.AdminKey != "",
	}
}

type switchBody struct {
	Profile string `json:"profile"`
	Force   bool   `json:"force"`
}

type versionsResponse struct {
	Active    string   `json:"active"`
	Installed []string `json:"installed"`
}

type downloadBody struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

type logsResponse struct {
	Lines []string `json:"lines"`
}

type eventsResponse struct {
	Events []logstore.Event `json:"events"`
	Next   int              `json:"next"`
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
