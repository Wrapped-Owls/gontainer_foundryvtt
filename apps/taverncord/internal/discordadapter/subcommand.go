package discordadapter

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

// Subcommand names and option keys, shared so a spec, its handler and the tests
// cannot drift apart.
const (
	subList        = "list"
	subStatus      = "status"
	subVersions    = "versions"
	subLogs        = "logs"
	subSwitch      = "switch"
	subRestart     = "restart"
	subDownload    = "download"
	subProfileEdit = "profile-edit"
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

// String returns the string value of the named option, or "" if absent.
func (m OptionMap) String(key string) string {
	if opt, ok := m[key]; ok {
		return opt.StringValue()
	}
	return ""
}

// Bool returns the boolean value of the named option, or false if absent.
func (m OptionMap) Bool(key string) bool {
	if opt, ok := m[key]; ok {
		return opt.BoolValue()
	}
	return false
}

// Int returns the integer value of the named option, or 0 if absent.
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

// suggester completes one option's value from what has been typed so far.
type suggester func(ctx context.Context, typed string) []string

// subCommand is the declarative form of a subcommand: the Discord spec is data,
// the behaviour is a function, and each completable option names the suggester
// that fills it. Declaring them as data is what keeps Autocomplete answering only
// for options the subcommand actually owns.
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

// parseInvocation descends to the invoked subcommand. Discord nests the real options
// one level down, so the leaf is read by shape rather than by a fixed index.
func parseInvocation(data discordgo.ApplicationCommandInteractionData) (invocation, bool) {
	opts := data.Options
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
