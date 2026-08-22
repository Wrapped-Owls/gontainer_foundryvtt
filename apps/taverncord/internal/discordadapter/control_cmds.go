package discordadapter

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

// ControlCommands groups the subcommands that can interrupt a live session, which
// is why every one of them carries the force flag.
func ControlCommands(cmds *command.ProfileCommands) []subCommand {
	return []subCommand{
		{
			name:        subSwitch,
			description: "Switch to a Foundry VTT profile",
			options: []*discordgo.ApplicationCommandOption{
				required(completedOption(optionName, "Profile name to activate")),
				forceOption("Switch even if players are currently online"),
			},
			handle: func(ctx context.Context, opts OptionMap, r command.Responder) error {
				return cmds.Switch(ctx, r, opts.String(optionName), opts.Bool(optionForce))
			},
			suggest: map[string]suggester{optionName: cmds.SuggestProfiles},
		},
		{
			name:        subRestart,
			description: "Restart Foundry VTT on the current profile",
			options: []*discordgo.ApplicationCommandOption{
				forceOption("Restart even if players are currently online"),
			},
			handle: func(ctx context.Context, opts OptionMap, r command.Responder) error {
				return cmds.Restart(ctx, r, opts.Bool(optionForce))
			},
		},
		{
			name:        subDownload,
			description: "Download a Foundry VTT version",
			options: []*discordgo.ApplicationCommandOption{
				required(completedOption(optionVersion, "Version to download (e.g. 14.361.0)")),
				stringOption(optionURL, "Optional presigned download URL"),
			},
			handle: func(ctx context.Context, opts OptionMap, r command.Responder) error {
				return cmds.Download(ctx, r, opts.String(optionVersion), opts.String(optionURL))
			},
			suggest: map[string]suggester{optionVersion: cmds.SuggestVersions},
		},
		{
			name:        subProfileEdit,
			description: "Edit an existing Foundry profile",
			options: []*discordgo.ApplicationCommandOption{
				required(completedOption(optionName, "Profile to edit")),
				stringOption(optionLabel, "Display label"),
				completedOption(optionVersion, "Foundry version"),
				stringOption(optionWorld, "World to launch on start"),
			},
			handle: func(ctx context.Context, opts OptionMap, r command.Responder) error {
				name := opts.String(optionName)
				return cmds.EditProfile(ctx, r, name, command.ProfileInput{
					Name:    name,
					Label:   opts.String(optionLabel),
					Version: opts.String(optionVersion),
					World:   opts.String(optionWorld),
				})
			},
			suggest: map[string]suggester{
				optionName:    cmds.SuggestProfiles,
				optionVersion: cmds.SuggestVersions,
			},
		},
	}
}
