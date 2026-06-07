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

type FoundryClient interface {
	ListProfiles(ctx context.Context) (ProfilesData, error)
	Switch(ctx context.Context, name string) error
	Status(ctx context.Context) (StatusData, error)
}

type ProfilesData struct {
	Active   string
	Profiles []profile.Profile
}

type StatusData struct {
	Active  string
	Version string
}
