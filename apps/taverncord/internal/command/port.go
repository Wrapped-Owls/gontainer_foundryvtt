package command

import (
	"context"

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
	Status(ctx context.Context) (StatusData, error)
	Versions(ctx context.Context) (VersionsData, error)
	Download(ctx context.Context, version, url string) error
	GetProfile(ctx context.Context, name string) (ProfileInfo, error)
	CreateProfile(ctx context.Context, p ProfileInput) error
	UpdateProfile(ctx context.Context, name string, p ProfileInput) error
	DeleteProfile(ctx context.Context, name string) error
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
	Name         string
	Label        string
	DataPath     string
	Version      string
	World        string
	ManifestPath string
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
