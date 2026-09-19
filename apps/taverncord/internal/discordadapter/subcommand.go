package discordadapter

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

const (
	subList     = "list"
	subStatus   = "status"
	subSwitch   = "switch"
	subDownload = "download"
)

const (
	optionName    = "name"
	optionVersion = "version"
	optionForce   = "force"
	optionURL     = "url"
	optionTail    = "tail"
	optionLabel   = "label"
	optionWorld   = "world"
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

type suggester func(ctx context.Context, typed string) []string

type subCommand struct {
	name        string
	description string
	options     []*discordgo.ApplicationCommandOption
	handle      func(ctx context.Context, opts OptionMap, r command.Responder) error
	suggest     map[string]suggester
}

func (c subCommand) Spec() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        c.name,
		Description: c.description,
		Options:     c.options,
	}
}

func (c subCommand) Handle(ctx context.Context, opts OptionMap, r command.Responder) error {
	return c.handle(ctx, opts, r)
}

func (c subCommand) Autocomplete(ctx context.Context, focused, typed string) []string {
	suggest, canSuggest := c.suggest[focused]
	if !canSuggest {
		return nil
	}
	return suggest(ctx, typed)
}

type invocation struct {
	Name    string
	Options []*discordgo.ApplicationCommandInteractionDataOption
}

func parseInvocation(interaction discordgo.ApplicationCommandInteractionData) (invocation, bool) {
	opts := interaction.Options // Discord nests the real options one level down
	if len(opts) != 1 || opts[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		return invocation{}, false
	}
	return invocation{Name: opts[0].Name, Options: opts[0].Options}, true
}

func focusedOption(
	opts []*discordgo.ApplicationCommandInteractionDataOption,
) (name, typed string) {
	for _, opt := range opts {
		if opt.Focused {
			return opt.Name, opt.StringValue()
		}
	}
	return "", ""
}

func stringOption(name, description string) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        name,
		Description: description,
	}
}

func completedOption(name, description string) *discordgo.ApplicationCommandOption {
	opt := stringOption(name, description)
	opt.Autocomplete = true
	return opt
}

func required(opt *discordgo.ApplicationCommandOption) *discordgo.ApplicationCommandOption {
	opt.Required = true
	return opt
}

func forceOption(description string) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionBoolean,
		Name:        optionForce,
		Description: description,
	}
}
