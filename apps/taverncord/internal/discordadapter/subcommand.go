package discordadapter

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

type OptionMap map[string]*discordgo.ApplicationCommandInteractionDataOption

func (m OptionMap) String(key string) string {
	if opt, ok := m[key]; ok {
		return opt.StringValue()
	}
	return ""
}

func (m OptionMap) Bool(key string) bool {
	if opt, ok := m[key]; ok {
		return opt.BoolValue()
	}
	return false
}

func (m OptionMap) Int(key string) int {
	if opt, ok := m[key]; ok {
		return int(opt.IntValue())
	}
	return 0
}

func newOptionMap(opts []*discordgo.ApplicationCommandInteractionDataOption) OptionMap {
	m := make(OptionMap, len(opts))
	for _, o := range opts {
		m[o.Name] = o
	}
	return m
}

type SubCommand interface {
	Spec() *discordgo.ApplicationCommandOption
	Handle(ctx context.Context, opts OptionMap, r command.Responder) error
}
