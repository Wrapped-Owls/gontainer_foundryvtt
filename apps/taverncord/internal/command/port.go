// Package command contains the bot's core logic, free of any Discord library dependency.
package command

import (
	"context"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

// Responder abstracts sending a reply to a Discord interaction.
// Implemented by discordadapter.interactionContext.
type Responder interface {
	Send(ctx context.Context, content string, ephemeral bool) error
	// Edit replaces the content of the initial response already sent by Send.
	Edit(ctx context.Context, content string) error
}

// FoundryClient abstracts calls to the foundrymanager dashboard REST API.
// Implemented by foundryclient.Client.
type FoundryClient interface {
	ListProfiles(ctx context.Context) (ProfilesData, error)
	Switch(ctx context.Context, name string) error
	Status(ctx context.Context) (StatusData, error)
}

// ProfilesData is the response shape of GET /profiles.
type ProfilesData struct {
	Active   string
	Profiles []profile.Profile
}

// StatusData is the response shape of GET /status.
type StatusData struct {
	Active  string
	Version string
}
