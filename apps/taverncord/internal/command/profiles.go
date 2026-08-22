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

// EditProfile restarts Foundry when the edit touches the active profile's version
// or world, so the reply says so instead of letting the server drop unexplained.
func (pc *ProfileCommands) EditProfile(
	ctx context.Context,
	r Responder,
	name string,
	p ProfileInput,
) error {
	willRestart, before := pc.editRestartsServer(ctx, name, p)

	if err := pc.client.UpdateProfile(ctx, name, p); err != nil {
		pc.logger.Error("edit profile failed", "profile", name, "err", err)
		return r.Send(ctx, fmt.Sprintf("❌ Edit failed: %s", err.Error()), true)
	}
	if !willRestart {
		return r.Send(ctx, fmt.Sprintf("✅ Updated profile **%s**.", name), true)
	}

	if err := r.Send(ctx, fmt.Sprintf(
		"✅ Updated profile **%s** - it is the active one, so the server is restarting...", name,
	), false); err != nil {
		return err
	}
	isBack := cycled(before)
	return pc.confirm(ctx, r, fmt.Sprintf("Updated **%s**", name), func(data StatusData) bool {
		return isBack(data) && data.Active == name
	})
}

// editRestartsServer answers false on a failed lookup rather than promise a restart
// that may never happen.
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
