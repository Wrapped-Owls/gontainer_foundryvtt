package command

import (
	"context"
	"fmt"
	"strings"
)

// ShowProfile fetches a single profile and sends its (secret-free) detail.
func (pc *ProfileCommands) ShowProfile(ctx context.Context, r Responder, name string) error {
	info, err := pc.client.GetProfile(ctx, name)
	if err != nil {
		pc.logger.Error("get profile failed", "profile", name, "err", err)
		return r.Send(ctx, fmt.Sprintf("❌ %s", err.Error()), true)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "**Profile `%s`**\n", info.Name)
	writeField(&sb, "Label", info.Label)
	writeField(&sb, "Data path", info.DataPath)
	writeField(&sb, "Version", info.Version)
	writeField(&sb, "World", info.World)
	writeField(&sb, "Manifest", info.ManifestPath)
	admin := "no"
	if info.HasAdminKey {
		admin = "yes"
	}
	fmt.Fprintf(&sb, "**Admin key set:** %s", admin)
	return r.Send(ctx, sb.String(), true)
}

// EditProfile updates the non-empty fields of an existing profile.
func (pc *ProfileCommands) EditProfile(
	ctx context.Context,
	r Responder,
	name string,
	p ProfileInput,
) error {
	if err := pc.client.UpdateProfile(ctx, name, p); err != nil {
		pc.logger.Error("edit profile failed", "profile", name, "err", err)
		return r.Send(ctx, fmt.Sprintf("❌ Edit failed: %s", err.Error()), true)
	}
	return r.Send(ctx, fmt.Sprintf("✅ Updated profile **%s**.", name), true)
}

// writeField appends a labelled line only when the value is non-empty.
func writeField(sb *strings.Builder, label, value string) {
	if value != "" {
		fmt.Fprintf(sb, "**%s:** `%s`\n", label, value)
	}
}
