package discordadapter

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

type listCmd struct{ cmds *command.ProfileCommands }

func (c *listCmd) Spec() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "list",
		Description: "List all available Foundry VTT profiles",
	}
}

func (c *listCmd) Handle(ctx context.Context, _ OptionMap, r command.Responder) error {
	return c.cmds.List(ctx, r)
}

type switchCmd struct{ cmds *command.ProfileCommands }

func (c *switchCmd) Spec() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "switch",
		Description: "Switch to a Foundry VTT profile",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name",
				Description: "Profile name to activate",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "force",
				Description: "Switch even if players are currently online",
			},
		},
	}
}

func (c *switchCmd) Handle(ctx context.Context, opts OptionMap, r command.Responder) error {
	return c.cmds.Switch(ctx, r, opts.String("name"), interruptOf(opts))
}

type statusCmd struct{ cmds *command.ProfileCommands }

func (c *statusCmd) Spec() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "status",
		Description: "Show the currently active Foundry VTT profile and version",
	}
}

func (c *statusCmd) Handle(ctx context.Context, _ OptionMap, r command.Responder) error {
	return c.cmds.Status(ctx, r)
}

func ListCmd(cmds *command.ProfileCommands) SubCommand { return &listCmd{cmds: cmds} }

func SwitchCmd(cmds *command.ProfileCommands) SubCommand { return &switchCmd{cmds: cmds} }

func StatusCmd(cmds *command.ProfileCommands) SubCommand { return &statusCmd{cmds: cmds} }

func interruptOf(opts OptionMap) command.Interrupt {
	if opts.Bool("force") {
		return command.InterruptAlways
	}
	return command.InterruptWhenIdle
}
