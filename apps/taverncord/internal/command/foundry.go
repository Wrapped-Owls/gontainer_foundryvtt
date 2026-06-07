package command

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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
		fmt.Fprintf(&sb, "%s **%s** (`%s`)\n", marker, label, p.Name)
	}
	return r.Send(ctx, sb.String(), true)
}

// Switch requests a profile change and reports the outcome to the responder.
// It sends an immediate acknowledgement, then edits it once the HTTP call resolves.
func (pc *ProfileCommands) Switch(ctx context.Context, r Responder, name string) error {
	if err := r.Send(ctx, fmt.Sprintf("⏳ Switching to profile **%s**…", name), false); err != nil {
		return err
	}
	if err := pc.client.Switch(ctx, name); err != nil {
		pc.logger.Error("switch profile failed", "profile", name, "err", err)
		return r.Edit(ctx, fmt.Sprintf("❌ Switch failed: %s", err.Error()))
	}
	return r.Edit(ctx, fmt.Sprintf("✅ Switched to **%s** — server is restarting.", name))
}

// Status fetches the current active profile and version and sends it to the responder.
func (pc *ProfileCommands) Status(ctx context.Context, r Responder) error {
	data, err := pc.client.Status(ctx)
	if err != nil {
		pc.logger.Error("status failed", "err", err)
		return r.Send(ctx, "Failed to fetch status from Foundry.", true)
	}
	msg := fmt.Sprintf("**Active profile:** `%s`\n**Version:** `%s`", data.Active, data.Version)
	return r.Send(ctx, msg, true)
}
