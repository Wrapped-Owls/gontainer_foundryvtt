// Package discordadapter bridges discordgo and the command core.
package discordadapter

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

// OptionMap wraps the raw discordgo option slice for convenient value access.
type OptionMap map[string]*discordgo.ApplicationCommandInteractionDataOption

// String returns the string value of the named option, or "" if absent.
func (m OptionMap) String(key string) string {
	if opt, ok := m[key]; ok {
		return opt.StringValue()
	}
	return ""
}

// newOptionMap builds an OptionMap from a slice of interaction data options.
func newOptionMap(opts []*discordgo.ApplicationCommandInteractionDataOption) OptionMap {
	m := make(OptionMap, len(opts))
	for _, o := range opts {
		m[o.Name] = o
	}
	return m
}

// SubCommand is the self-describing command unit, analogous to http.Handler.
// Spec() provides the Discord registration metadata; Handle() executes the logic.
type SubCommand interface {
	Spec() *discordgo.ApplicationCommandOption
	Handle(ctx context.Context, opts OptionMap, r command.Responder) error
}
