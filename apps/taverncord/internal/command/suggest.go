package command

import (
	"context"
	"slices"
	"strings"
)

// maxSuggestions is Discord's hard cap on autocomplete choices per response.
const maxSuggestions = 25

// SuggestProfiles returns nil on failure: autocomplete has nowhere to show a message.
func (pc *ProfileCommands) SuggestProfiles(ctx context.Context, typed string) []string {
	data, err := pc.client.ListProfiles(ctx)
	if err != nil {
		pc.logger.Warn("profile suggestions unavailable", "err", err)
		return nil
	}
	names := make([]string, 0, len(data.Profiles))
	for _, p := range data.Profiles {
		names = append(names, p.Name)
	}
	return matching(names, typed)
}

func (pc *ProfileCommands) SuggestVersions(ctx context.Context, typed string) []string {
	data, err := pc.client.Versions(ctx)
	if err != nil {
		pc.logger.Warn("version suggestions unavailable", "err", err)
		return nil
	}
	return matching(data.Installed, typed)
}

func matching(candidates []string, typed string) []string {
	typed = strings.ToLower(strings.TrimSpace(typed))
	matches := slices.DeleteFunc(slices.Clone(candidates), func(candidate string) bool {
		return !strings.Contains(strings.ToLower(candidate), typed)
	})
	return matches[:min(len(matches), maxSuggestions)]
}
