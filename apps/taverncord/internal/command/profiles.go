package command

import (
	"context"
	"fmt"
	"strings"
)

func (pc *ProfileCommands) ShowProfile(ctx context.Context, r Responder, name string) error {
	shown, err := pc.client.GetProfile(ctx, name)
	if err != nil {
		pc.logger.Error("get profile failed", "profile", name, "err", err)
		return r.Send(ctx, fmt.Sprintf("❌ %s", err.Error()), Private)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "**Profile `%s`**\n", shown.Name)
	writeField(&sb, "Label", shown.Label)
	writeField(&sb, "Data path", shown.DataPath)
	writeField(&sb, "Version", shown.Version)
	writeField(&sb, "World", shown.World)
	writeField(&sb, "Manifest", shown.ManifestPath)
	admin := "no"
	if shown.HasAdminKey {
		admin = "yes"
	}
	fmt.Fprintf(&sb, "**Admin key set:** %s", admin)
	return r.Send(ctx, sb.String(), Private)
}

func (pc *ProfileCommands) EditProfile(
	ctx context.Context,
	r Responder,
	name string,
	p ProfileInput,
) error {
	willRestart, before := pc.editRestartsServer(ctx, name, p)

	if err := pc.client.UpdateProfile(ctx, name, p); err != nil {
		pc.logger.Error("edit profile failed", "profile", name, "err", err)
		return r.Send(ctx, fmt.Sprintf("❌ Edit failed: %s", err.Error()), Private)
	}
	if !willRestart {
		return r.Send(ctx, fmt.Sprintf("✅ Updated profile **%s**.", name), Private)
	}

	if err := r.Send(ctx, fmt.Sprintf(
		"✅ Updated profile **%s** - it is the active one, so the server is restarting...", name,
	), Public); err != nil {
		return err
	}
	isBack := cycled(before)
	return pc.confirm(ctx, r, fmt.Sprintf("Updated **%s**", name), func(status StatusData) bool {
		return isBack(status) && status.Active == name
	})
}

func (pc *ProfileCommands) editRestartsServer(
	ctx context.Context,
	name string,
	p ProfileInput,
) (bool, StatusData) {
	status, err := pc.client.Status(ctx)
	if err != nil {
		pc.logger.Warn("status before edit failed", "err", err)
		return false, status
	}
	if status.Active != name {
		return false, status
	}
	current, err := pc.client.GetProfile(ctx, name)
	if err != nil {
		return false, status
	}
	restarts := (p.Version != "" && p.Version != current.Version) ||
		(p.World != "" && p.World != current.World)
	return restarts, status
}

func writeField(sb *strings.Builder, label, value string) {
	if value != "" {
		fmt.Fprintf(sb, "**%s:** `%s`\n", label, value)
	}
}
