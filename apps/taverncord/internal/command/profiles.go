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

func (pc *ProfileCommands) CreateProfile(ctx context.Context, r Responder, p ProfileInput) error {
	if err := pc.client.CreateProfile(ctx, p); err != nil {
		pc.logger.Error("create profile failed", "profile", p.Name, "err", err)
		return r.Send(ctx, fmt.Sprintf("❌ Create failed: %s", err.Error()), Private)
	}
	return r.Send(ctx, fmt.Sprintf("✅ Created profile **%s**.", p.Name), Private)
}

func (pc *ProfileCommands) EditProfile(
	ctx context.Context,
	r Responder,
	name string,
	p ProfileInput,
) error {
	if err := pc.client.UpdateProfile(ctx, name, p); err != nil {
		pc.logger.Error("edit profile failed", "profile", name, "err", err)
		return r.Send(ctx, fmt.Sprintf("❌ Edit failed: %s", err.Error()), Private)
	}
	return r.Send(ctx, fmt.Sprintf("✅ Updated profile **%s**.", name), Private)
}

func (pc *ProfileCommands) DeleteProfile(ctx context.Context, r Responder, name string) error {
	if err := pc.client.DeleteProfile(ctx, name); err != nil {
		pc.logger.Error("delete profile failed", "profile", name, "err", err)
		return r.Send(ctx, fmt.Sprintf("❌ Delete failed: %s", err.Error()), Private)
	}
	return r.Send(ctx, fmt.Sprintf("✅ Deleted profile **%s**.", name), Private)
}

func writeField(sb *strings.Builder, label, value string) {
	if value != "" {
		fmt.Fprintf(sb, "**%s:** `%s`\n", label, value)
	}
}
