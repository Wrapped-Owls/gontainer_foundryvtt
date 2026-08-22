package command

import (
	"context"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

// Responder abstracts a reply; implemented by discordadapter.interactionContext.
type Responder interface {
	Send(ctx context.Context, content string, ephemeral bool) error
	// Edit replaces the content of the initial response already sent by Send.
	Edit(ctx context.Context, content string) error
}

// FoundryClient abstracts the dashboard REST API; implemented by foundryclient.Client.
type FoundryClient interface {
	ListProfiles(ctx context.Context) (ProfilesData, error)
	Switch(ctx context.Context, name string, force bool) error
	Restart(ctx context.Context, force bool) error
	Status(ctx context.Context) (StatusData, error)
	Versions(ctx context.Context) (VersionsData, error)
	Download(ctx context.Context, version, url string) error
	GetProfile(ctx context.Context, name string) (ProfileInfo, error)
	UpdateProfile(ctx context.Context, name string, p ProfileInput) error
	Logs(ctx context.Context, tail int) (LogsData, error)
	Events(ctx context.Context, since int) (EventsData, error)
}

// LogsData is the response shape of GET /logs.
type LogsData struct {
	Lines []string
}

// EventsData is the response shape of GET /events: detected error/crash events
// plus the cursor to resume polling from.
type EventsData struct {
	Events []EventItem
	Next   int
}

type EventItem struct {
	Time    time.Time
	Kind    string
	Message string
}

// ProfileInfo is the non-secret view of a profile returned by GET /profiles/{name}.
type ProfileInfo struct {
	Name         string
	Label        string
	DataPath     string
	Version      string
	World        string
	ManifestPath string
	HasAdminKey  bool
}

// ProfileInput carries the editable fields of a profile. The data location,
// manifest path and admin secrets are intentionally excluded so an edit can
// never repoint a profile at another disk or set credentials from Discord.
type ProfileInput struct {
	Name    string
	Label   string
	Version string
	World   string
}

// ProfilesData is the response shape of GET /profiles.
type ProfilesData struct {
	Active   string
	Profiles []profile.Profile
}

// VersionsData is the response shape of GET /versions.
type VersionsData struct {
	Active    string
	Installed []string
}

// StatusData is the response shape of GET /status: the active profile plus the
// live status of the running Foundry server (zero-valued when offline).
type StatusData struct {
	Active        string
	Version       string
	Online        bool
	WorldActive   bool
	World         string
	System        string
	SystemVersion string
	Users         int
	UptimeMS      int64
}
