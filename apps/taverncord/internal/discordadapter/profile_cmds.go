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

type dataPathRequirement string

func editableOptions(dataPathRequired dataPathRequirement) []*discordgo.ApplicationCommandOption {
	const dataPathMandatory dataPathRequirement = "mandatory"

	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "datapath",
			Description: "Foundry data directory for this profile",
			Required:    dataPathRequired == dataPathMandatory,
		},
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
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "manifest",
			Description: "Patch manifest path",
		},
	}
}

func profileInput(opts OptionMap, name string) command.ProfileInput {
	return command.ProfileInput{
		Name:         name,
		Label:        opts.String("label"),
		DataPath:     opts.String("datapath"),
		Version:      opts.String("version"),
		World:        opts.String("world"),
		ManifestPath: opts.String("manifest"),
	}
}

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

type profileCreateCmd struct{ cmds *command.ProfileCommands }

func (c *profileCreateCmd) Spec() *discordgo.ApplicationCommandOption {
	const dataPathMandatory dataPathRequirement = "mandatory"

	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "profile-create",
		Description: "Create a new Foundry profile",
		Options: append(
			[]*discordgo.ApplicationCommandOption{nameOption("New profile name")},
			editableOptions(dataPathMandatory)...,
		),
	}
}

func (c *profileCreateCmd) Handle(ctx context.Context, opts OptionMap, r command.Responder) error {
	return c.cmds.CreateProfile(ctx, r, profileInput(opts, opts.String("name")))
}

type profileEditCmd struct{ cmds *command.ProfileCommands }

func (c *profileEditCmd) Spec() *discordgo.ApplicationCommandOption {
	const dataPathOptional dataPathRequirement = "optional"

	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "profile-edit",
		Description: "Edit an existing Foundry profile",
		Options: append(
			[]*discordgo.ApplicationCommandOption{nameOption("Profile to edit")},
			editableOptions(dataPathOptional)...,
		),
	}
}

func (c *profileEditCmd) Handle(ctx context.Context, opts OptionMap, r command.Responder) error {
	name := opts.String("name")
	return c.cmds.EditProfile(ctx, r, name, profileInput(opts, name))
}

type profileDeleteCmd struct{ cmds *command.ProfileCommands }

func (c *profileDeleteCmd) Spec() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "profile-delete",
		Description: "Delete a Foundry profile",
		Options:     []*discordgo.ApplicationCommandOption{nameOption("Profile to delete")},
	}
}

func (c *profileDeleteCmd) Handle(ctx context.Context, opts OptionMap, r command.Responder) error {
	return c.cmds.DeleteProfile(ctx, r, opts.String("name"))
}

func ProfileShowCmd(cmds *command.ProfileCommands) SubCommand { return &profileShowCmd{cmds: cmds} }

func ProfileCreateCmd(
	cmds *command.ProfileCommands,
) SubCommand {
	return &profileCreateCmd{cmds: cmds}
}

func ProfileEditCmd(cmds *command.ProfileCommands) SubCommand { return &profileEditCmd{cmds: cmds} }

func ProfileDeleteCmd(
	cmds *command.ProfileCommands,
) SubCommand {
	return &profileDeleteCmd{cmds: cmds}
}
