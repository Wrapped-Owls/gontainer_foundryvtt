package discordadapter

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

func nameOption(desc string) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "name",
		Description: desc,
		Required:    true,
	}
}

// editableOptions are the profile fields editable from Discord. The data
// location, manifest path and admin secrets are intentionally absent so an edit
// cannot repoint a profile at another disk or set credentials from a channel.
func editableOptions() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "label",
			Description: "Display label",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "version",
			Description: "Foundry version",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "world",
			Description: "World to launch on start",
		},
	}
}

func profileInput(opts OptionMap, name string) command.ProfileInput {
	return command.ProfileInput{
		Name:    name,
		Label:   opts.String("label"),
		Version: opts.String("version"),
		World:   opts.String("world"),
	}
}

// profileShowCmd handles /foundry profile-show.
type profileShowCmd struct{ cmds *command.ProfileCommands }

func (c *profileShowCmd) Spec() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "profile-show",
		Description: "Show a profile's configuration",
		Options:     []*discordgo.ApplicationCommandOption{nameOption("Profile name")},
	}
}

func (c *profileShowCmd) Handle(ctx context.Context, opts OptionMap, r command.Responder) error {
	return c.cmds.ShowProfile(ctx, r, opts.String("name"))
}

// profileEditCmd handles /foundry profile-edit.
type profileEditCmd struct{ cmds *command.ProfileCommands }

func (c *profileEditCmd) Spec() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "profile-edit",
		Description: "Edit an existing Foundry profile",
		Options: append(
			[]*discordgo.ApplicationCommandOption{nameOption("Profile to edit")},
			editableOptions()...,
		),
	}
}

func (c *profileEditCmd) Handle(ctx context.Context, opts OptionMap, r command.Responder) error {
	name := opts.String("name")
	return c.cmds.EditProfile(ctx, r, name, profileInput(opts, name))
}

// ProfileShowCmd returns a SubCommand for /foundry profile-show.
func ProfileShowCmd(cmds *command.ProfileCommands) SubCommand { return &profileShowCmd{cmds: cmds} }

// ProfileEditCmd returns a SubCommand for /foundry profile-edit.
func ProfileEditCmd(cmds *command.ProfileCommands) SubCommand { return &profileEditCmd{cmds: cmds} }
