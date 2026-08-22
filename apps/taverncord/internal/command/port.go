package command

import (
	"context"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

type Visibility string

const (
	Private Visibility = "private"
	Public  Visibility = "public"
)

type Responder interface {
	Send(ctx context.Context, content string, visibility Visibility) error
	Edit(ctx context.Context, content string) error
}

type Interrupt string

const (
	InterruptWhenIdle Interrupt = "when-idle" // the manager refuses with 409 while players are online
	InterruptAlways   Interrupt = "always"
)

type FoundryClient interface {
	ListProfiles(ctx context.Context) (ProfilesData, error)
	Switch(ctx context.Context, name string, interrupt Interrupt) error
	Restart(ctx context.Context, interrupt Interrupt) error
	Status(ctx context.Context) (StatusData, error)
	Versions(ctx context.Context) (VersionsData, error)
	Download(ctx context.Context, version, url string) error
	GetProfile(ctx context.Context, name string) (ProfileInfo, error)
	UpdateProfile(ctx context.Context, name string, p ProfileInput) error
	Logs(ctx context.Context, tail int) (LogsData, error)
	Events(ctx context.Context, since int) (EventsData, error)
}

type LogsData struct {
	Lines []string
}

type EventsData struct {
	Events []EventItem
	Next   int
}

type EventItem struct {
	Time    time.Time
	Kind    string
	Message string
}

type ProfileInfo struct {
	Name         string
	Label        string
	DataPath     string
	Version      string
	World        string
	ManifestPath string
	HasAdminKey  bool
}

type ProfileInput struct {
	Name    string
	Label   string
	Version string
	World   string
}

type ProfilesData struct {
	Active   string
	Profiles []profile.Profile
}

type VersionsData struct {
	Active    string
	Installed []string
}

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
