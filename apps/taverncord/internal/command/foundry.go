package command

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// ProfileCommands implements the /foundry subcommand logic.
// It has no dependency on any Discord library.
type ProfileCommands struct {
	client FoundryClient
	logger *slog.Logger
}

// New creates a ProfileCommands using the provided client and logger.
func New(client FoundryClient, logger *slog.Logger) *ProfileCommands {
	return &ProfileCommands{client: client, logger: logger}
}

// List fetches all profiles and sends a formatted list to the responder.
func (pc *ProfileCommands) List(ctx context.Context, r Responder) error {
	data, err := pc.client.ListProfiles(ctx)
	if err != nil {
		pc.logger.Error("list profiles failed", "err", err)
		return r.Send(ctx, "Failed to fetch profiles from Foundry.", true)
	}
	if len(data.Profiles) == 0 {
		return r.Send(ctx, "No profiles configured.", true)
	}

	var sb strings.Builder
	sb.WriteString("**Foundry Profiles**\n")
	for _, p := range data.Profiles {
		label := p.Label
		if label == "" {
			label = p.Name
		}
		marker := "○"
		if p.Name == data.Active {
			marker = "▶"
		}
		fmt.Fprintf(&sb, "%s **%s** (`%s`)", marker, label, p.Name)
		if detail := profileSummary(p.Version, p.World); detail != "" {
			sb.WriteString(" — " + detail)
		}
		sb.WriteByte('\n')
	}
	sb.WriteString("_Use `/foundry list name:<profile>` for full details._")
	return r.Send(ctx, sb.String(), true)
}

// profileSummary renders the compact "version • world" tail shown per list row,
// omitting whichever field is unset (empty when both are).
func profileSummary(version, world string) string {
	var parts []string
	if version != "" {
		parts = append(parts, "v"+version)
	}
	if world != "" {
		parts = append(parts, "world `"+world+"`")
	}
	return strings.Join(parts, " • ")
}

// Switch requests a profile change and reports the outcome to the responder.
// It sends an immediate acknowledgement, then edits it once the HTTP call resolves.
// When force is false the switch is refused if players are connected.
func (pc *ProfileCommands) Switch(ctx context.Context, r Responder, name string, force bool) error {
	if err := r.Send(ctx, fmt.Sprintf("⏳ Switching to profile **%s**…", name), false); err != nil {
		return err
	}
	if err := pc.client.Switch(ctx, name, force); err != nil {
		pc.logger.Error("switch profile failed", "profile", name, "err", err)
		return r.Edit(ctx, fmt.Sprintf("❌ Switch failed: %s", err.Error()))
	}
	return r.Edit(ctx, fmt.Sprintf("✅ Switched to **%s** — server is restarting.", name))
}

// logsCharBudget stays well under Discord's 2000-character message cap to leave
// room for the code-fence wrapper.
const logsCharBudget = 1800

// Logs fetches the most recent Foundry log lines and sends them in a code block.
func (pc *ProfileCommands) Logs(ctx context.Context, r Responder, tail int) error {
	data, err := pc.client.Logs(ctx, tail)
	if err != nil {
		pc.logger.Error("fetch logs failed", "err", err)
		return r.Send(ctx, "Failed to fetch logs from Foundry.", true)
	}
	if len(data.Lines) == 0 {
		return r.Send(ctx, "No logs captured yet.", true)
	}
	body := strings.Join(data.Lines, "\n")
	if len(body) > logsCharBudget {
		body = "…" + body[len(body)-logsCharBudget:]
	}
	return r.Send(ctx, "```\n"+body+"\n```", true)
}

// Versions lists the installed Foundry versions and the active one.
func (pc *ProfileCommands) Versions(ctx context.Context, r Responder) error {
	data, err := pc.client.Versions(ctx)
	if err != nil {
		pc.logger.Error("list versions failed", "err", err)
		return r.Send(ctx, "Failed to fetch versions from Foundry.", true)
	}
	if len(data.Installed) == 0 {
		return r.Send(ctx, "No Foundry versions installed.", true)
	}

	var sb strings.Builder
	sb.WriteString("**Installed Foundry versions**\n")
	for _, v := range data.Installed {
		marker := "○"
		if v == data.Active {
			marker = "▶"
		}
		fmt.Fprintf(&sb, "%s `%s`\n", marker, v)
	}
	return r.Send(ctx, sb.String(), true)
}

// Download acquires a Foundry version through the manager. It acknowledges
// immediately, then edits the reply once the (possibly slow) download resolves.
func (pc *ProfileCommands) Download(ctx context.Context, r Responder, version, url string) error {
	if err := r.Send(ctx, fmt.Sprintf("⏳ Downloading Foundry **%s**…", version), true); err != nil {
		return err
	}
	if err := pc.client.Download(ctx, version, url); err != nil {
		pc.logger.Error("download version failed", "version", version, "err", err)
		return r.Edit(ctx, fmt.Sprintf("❌ Download failed: %s", err.Error()))
	}
	return r.Edit(ctx, fmt.Sprintf("✅ Foundry **%s** is ready to use.", version))
}

// Status fetches the active profile and the live Foundry server status and sends
// a formatted summary to the responder.
func (pc *ProfileCommands) Status(ctx context.Context, r Responder) error {
	data, err := pc.client.Status(ctx)
	if err != nil {
		pc.logger.Error("status failed", "err", err)
		return r.Send(ctx, "Failed to fetch status from Foundry.", true)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "**Active profile:** `%s`\n**Version:** `%s`\n", data.Active, data.Version)
	if !data.Online {
		sb.WriteString("**Server:** ⚫ offline")
		return r.Send(ctx, sb.String(), true)
	}
	sb.WriteString("**Server:** 🟢 online\n")
	if data.WorldActive && data.World != "" {
		fmt.Fprintf(&sb, "**World:** `%s`\n", data.World)
		if data.System != "" {
			fmt.Fprintf(&sb, "**System:** `%s %s`\n", data.System, data.SystemVersion)
		}
	} else {
		sb.WriteString("**World:** none active (setup screen)\n")
	}
	fmt.Fprintf(&sb, "**Users online:** %d", data.Users)
	if data.UptimeMS > 0 {
		uptime := (time.Duration(data.UptimeMS) * time.Millisecond).Round(time.Second)
		fmt.Fprintf(&sb, "\n**Uptime:** %s", uptime)
	}
	return r.Send(ctx, sb.String(), true)
}
