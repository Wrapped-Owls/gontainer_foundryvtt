package command

import (
	"context"
	"slices"
	"strings"
)

const maxSuggestions = 25

func (pc *ProfileCommands) SuggestProfiles(ctx context.Context, typed string) []string {
	listing, err := pc.client.ListProfiles(ctx)
	if err != nil {
		pc.logger.Warn("profile suggestions unavailable", "err", err)
		return nil
	}
	names := make([]string, 0, len(listing.Profiles))
	for _, p := range listing.Profiles {
		names = append(names, p.Name)
	}
	return matching(names, typed)
}

func (pc *ProfileCommands) SuggestVersions(ctx context.Context, typed string) []string {
	versions, err := pc.client.Versions(ctx)
	if err != nil {
		pc.logger.Warn("version suggestions unavailable", "err", err)
		return nil
	}
	return matching(versions.Installed, typed)
}

func matching(candidates []string, typed string) []string {
	typed = strings.ToLower(strings.TrimSpace(typed))
	matches := slices.DeleteFunc(slices.Clone(candidates), func(candidate string) bool {
		return !strings.Contains(strings.ToLower(candidate), typed)
	})
	return matches[:min(len(matches), maxSuggestions)]
}
