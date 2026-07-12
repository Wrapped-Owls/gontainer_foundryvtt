package command

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type ProfileCommands struct {
	client FoundryClient
	logger *slog.Logger
}

func New(client FoundryClient, logger *slog.Logger) *ProfileCommands {
	return &ProfileCommands{client: client, logger: logger}
}

func (pc *ProfileCommands) List(ctx context.Context, r Responder) error {
	listing, err := pc.client.ListProfiles(ctx)
	if err != nil {
		pc.logger.Error("list profiles failed", "err", err)
		return r.Send(ctx, "Failed to fetch profiles from Foundry.", Private)
	}
	if len(listing.Profiles) == 0 {
		return r.Send(ctx, "No profiles configured.", Private)
	}

	var sb strings.Builder
	sb.WriteString("**Foundry Profiles**\n")
	for _, p := range listing.Profiles {
		label := p.Label
		if label == "" {
			label = p.Name
		}
		marker := "○"
		if p.Name == listing.Active {
			marker = "▶"
		}
		fmt.Fprintf(&sb, "%s **%s** (`%s`)\n", marker, label, p.Name)
	}
	return r.Send(ctx, sb.String(), Private)
}

func (pc *ProfileCommands) Switch(
	ctx context.Context,
	r Responder,
	name string,
	interrupt Interrupt,
) error {
	if err := r.Send(
		ctx,
		fmt.Sprintf("⏳ Switching to profile **%s**...", name),
		Public,
	); err != nil {
		return err
	}
	if err := pc.client.Switch(ctx, name, interrupt); err != nil {
		pc.logger.Error("switch profile failed", "profile", name, "err", err)
		return r.Edit(ctx, fmt.Sprintf("❌ Switch failed: %s", err.Error()))
	}
	return r.Edit(ctx, fmt.Sprintf("✅ Switched to **%s**  -  server is restarting.", name))
}

func (pc *ProfileCommands) Logs(ctx context.Context, r Responder, tail int) error {
	const logsCharBudget = 1800
	logs, err := pc.client.Logs(ctx, tail)
	if err != nil {
		pc.logger.Error("fetch logs failed", "err", err)
		return r.Send(ctx, "Failed to fetch logs from Foundry.", Private)
	}
	if len(logs.Lines) == 0 {
		return r.Send(ctx, "No logs captured yet.", Private)
	}
	body := strings.Join(logs.Lines, "\n")
	if len(body) > logsCharBudget {
		body = "..." + body[len(body)-logsCharBudget:]
	}
	return r.Send(ctx, "```\n"+body+"\n```", Private)
}

func (pc *ProfileCommands) Versions(ctx context.Context, r Responder) error {
	versions, err := pc.client.Versions(ctx)
	if err != nil {
		pc.logger.Error("list versions failed", "err", err)
		return r.Send(ctx, "Failed to fetch versions from Foundry.", Private)
	}
	if len(versions.Installed) == 0 {
		return r.Send(ctx, "No Foundry versions installed.", Private)
	}

	var sb strings.Builder
	sb.WriteString("**Installed Foundry versions**\n")
	for _, v := range versions.Installed {
		marker := "○"
		if v == versions.Active {
			marker = "▶"
		}
		fmt.Fprintf(&sb, "%s `%s`\n", marker, v)
	}
	return r.Send(ctx, sb.String(), Private)
}

func (pc *ProfileCommands) Download(ctx context.Context, r Responder, version, url string) error {
	if err := r.Send(
		ctx,
		fmt.Sprintf("⏳ Downloading Foundry **%s**...", version),
		Private,
	); err != nil {
		return err
	}
	if err := pc.client.Download(ctx, version, url); err != nil {
		pc.logger.Error("download version failed", "version", version, "err", err)
		return r.Edit(ctx, fmt.Sprintf("❌ Download failed: %s", err.Error()))
	}
	return r.Edit(ctx, fmt.Sprintf("✅ Foundry **%s** is ready to use.", version))
}

func (pc *ProfileCommands) Status(ctx context.Context, r Responder) error {
	status, err := pc.client.Status(ctx)
	if err != nil {
		pc.logger.Error("status failed", "err", err)
		return r.Send(ctx, "Failed to fetch status from Foundry.", Private)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "**Active profile:** `%s`\n**Version:** `%s`\n", status.Active, status.Version)
	if !status.Online {
		sb.WriteString("**Server:** ⚫ offline")
		return r.Send(ctx, sb.String(), Private)
	}
	sb.WriteString("**Server:** 🟢 online\n")
	if status.WorldActive && status.World != "" {
		fmt.Fprintf(&sb, "**World:** `%s`\n", status.World)
		if status.System != "" {
			fmt.Fprintf(&sb, "**System:** `%s %s`\n", status.System, status.SystemVersion)
		}
	} else {
		sb.WriteString("**World:** none active (setup screen)\n")
	}
	fmt.Fprintf(&sb, "**Users online:** %d", status.Users)
	if status.UptimeMS > 0 {
		uptime := (time.Duration(status.UptimeMS) * time.Millisecond).Round(time.Second)
		fmt.Fprintf(&sb, "\n**Uptime:** %s", uptime)
	}
	return r.Send(ctx, sb.String(), Private)
}
