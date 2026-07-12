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
