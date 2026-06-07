package command

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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

func (pc *ProfileCommands) Switch(ctx context.Context, r Responder, name string) error {
	if err := r.Send(
		ctx,
		fmt.Sprintf("⏳ Switching to profile **%s**...", name),
		Public,
	); err != nil {
		return err
	}
	if err := pc.client.Switch(ctx, name); err != nil {
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
	msg := fmt.Sprintf("**Active profile:** `%s`\n**Version:** `%s`", status.Active, status.Version)
	return r.Send(ctx, msg, Private)
}
