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
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "force",
				Description: "Switch even if players are currently online",
			},
		},
	}
}

func (c *switchCmd) Handle(ctx context.Context, opts OptionMap, r command.Responder) error {
	return c.cmds.Switch(ctx, r, opts.String("name"), opts.Bool("force"))
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

// versionsCmd handles /foundry versions.
type versionsCmd struct{ cmds *command.ProfileCommands }

func (c *versionsCmd) Spec() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "versions",
		Description: "List installed Foundry VTT versions",
	}
}

func (c *versionsCmd) Handle(ctx context.Context, _ OptionMap, r command.Responder) error {
	return c.cmds.Versions(ctx, r)
}

// downloadCmd handles /foundry download.
type downloadCmd struct{ cmds *command.ProfileCommands }

func (c *downloadCmd) Spec() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "download",
		Description: "Download a Foundry VTT version",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "version",
				Description: "Version to download (e.g. 14.361.0)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "Optional presigned download URL",
			},
		},
	}
}

func (c *downloadCmd) Handle(ctx context.Context, opts OptionMap, r command.Responder) error {
	return c.cmds.Download(ctx, r, opts.String("version"), opts.String("url"))
}

// logsCmd handles /foundry logs.
type logsCmd struct{ cmds *command.ProfileCommands }

func (c *logsCmd) Spec() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "logs",
		Description: "Show the most recent Foundry log lines",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "tail",
				Description: "How many lines to show (default 20)",
			},
		},
	}
}

func (c *logsCmd) Handle(ctx context.Context, opts OptionMap, r command.Responder) error {
	tail := opts.Int("tail")
	if tail == 0 {
		tail = 20
	}
	return c.cmds.Logs(ctx, r, tail)
}

// Constructor functions — used in main.go composition root.

// ListCmd returns a SubCommand for /foundry list.
func ListCmd(cmds *command.ProfileCommands) SubCommand { return &listCmd{cmds: cmds} }

// SwitchCmd returns a SubCommand for /foundry switch.
func SwitchCmd(cmds *command.ProfileCommands) SubCommand { return &switchCmd{cmds: cmds} }

// StatusCmd returns a SubCommand for /foundry status.
func StatusCmd(cmds *command.ProfileCommands) SubCommand { return &statusCmd{cmds: cmds} }

// VersionsCmd returns a SubCommand for /foundry versions.
func VersionsCmd(cmds *command.ProfileCommands) SubCommand { return &versionsCmd{cmds: cmds} }

// DownloadCmd returns a SubCommand for /foundry download.
func DownloadCmd(cmds *command.ProfileCommands) SubCommand { return &downloadCmd{cmds: cmds} }

// LogsCmd returns a SubCommand for /foundry logs.
func LogsCmd(cmds *command.ProfileCommands) SubCommand { return &logsCmd{cmds: cmds} }
