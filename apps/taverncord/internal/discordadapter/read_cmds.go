package discordadapter

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

const defaultLogTail = 20

func ReadCommands(cmds *command.ProfileCommands) []subCommand {
	return []subCommand{
		{
			name:        subList,
			description: "List Foundry VTT profiles, or show one when a name is given",
			options: []*discordgo.ApplicationCommandOption{
				completedOption(optionName, "Show this profile's full configuration"),
			},
			handle: func(ctx context.Context, opts OptionMap, r command.Responder) error {
				if name := opts.String(optionName); name != "" {
					return cmds.ShowProfile(ctx, r, name)
				}
				return cmds.List(ctx, r)
			},
			suggest: map[string]suggester{optionName: cmds.SuggestProfiles},
		},
		{
			name:        subStatus,
			description: "Show the currently active Foundry VTT profile and version",
			handle: func(ctx context.Context, _ OptionMap, r command.Responder) error {
				return cmds.Status(ctx, r)
			},
		},
		{
			name:        subVersions,
			description: "List installed Foundry VTT versions",
			handle: func(ctx context.Context, _ OptionMap, r command.Responder) error {
				return cmds.Versions(ctx, r)
			},
		},
		{
			name:        subLogs,
			description: "Show the most recent Foundry log lines",
			options: []*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        optionTail,
				Description: "How many lines to show (default 20)",
			}},
			handle: func(ctx context.Context, opts OptionMap, r command.Responder) error {
				tail := opts.Int(optionTail)
				if tail == 0 {
					tail = defaultLogTail
				}
				return cmds.Logs(ctx, r, tail)
			},
		},
	}
}
