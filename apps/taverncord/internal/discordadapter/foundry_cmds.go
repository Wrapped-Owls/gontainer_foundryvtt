package discordadapter

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

// listCmd handles /foundry list.
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

// switchCmd handles /foundry switch.
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
		},
	}
}

func (c *switchCmd) Handle(ctx context.Context, opts OptionMap, r command.Responder) error {
	return c.cmds.Switch(ctx, r, opts.String("name"))
}

// statusCmd handles /foundry status.
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

// Constructor functions — used in main.go composition root.

// ListCmd returns a SubCommand for /foundry list.
func ListCmd(cmds *command.ProfileCommands) SubCommand { return &listCmd{cmds: cmds} }

// SwitchCmd returns a SubCommand for /foundry switch.
func SwitchCmd(cmds *command.ProfileCommands) SubCommand { return &switchCmd{cmds: cmds} }

// StatusCmd returns a SubCommand for /foundry status.
func StatusCmd(cmds *command.ProfileCommands) SubCommand { return &statusCmd{cmds: cmds} }
