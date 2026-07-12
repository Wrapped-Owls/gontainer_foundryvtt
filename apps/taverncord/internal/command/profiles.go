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

// CreateProfile creates a new profile and reports the outcome.
func (pc *ProfileCommands) CreateProfile(ctx context.Context, r Responder, p ProfileInput) error {
	if err := pc.client.CreateProfile(ctx, p); err != nil {
		pc.logger.Error("create profile failed", "profile", p.Name, "err", err)
		return r.Send(ctx, fmt.Sprintf("❌ Create failed: %s", err.Error()), true)
	}
	return r.Send(ctx, fmt.Sprintf("✅ Created profile **%s**.", p.Name), true)
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

// DeleteProfile removes a profile and reports the outcome.
func (pc *ProfileCommands) DeleteProfile(ctx context.Context, r Responder, name string) error {
	if err := pc.client.DeleteProfile(ctx, name); err != nil {
		pc.logger.Error("delete profile failed", "profile", name, "err", err)
		return r.Send(ctx, fmt.Sprintf("❌ Delete failed: %s", err.Error()), true)
	}
	return r.Send(ctx, fmt.Sprintf("✅ Deleted profile **%s**.", name), true)
}

// writeField appends a labelled line only when the value is non-empty.
func writeField(sb *strings.Builder, label, value string) {
	if value != "" {
		fmt.Fprintf(sb, "**%s:** `%s`\n", label, value)
	}
}
