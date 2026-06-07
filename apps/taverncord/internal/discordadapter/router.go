package discordadapter

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// Router maps subcommand names to SubCommand implementations and dispatches
// Discord interactions, analogous to http.ServeMux.
type Router struct {
	name        string
	description string
	gmRoleID    string
	subs        map[string]SubCommand
	logger      *slog.Logger
}

// NewRouter creates a Router for a top-level slash command with the given name and description.
func NewRouter(name, description string, logger *slog.Logger) *Router {
	return &Router{
		name:        name,
		description: description,
		subs:        make(map[string]SubCommand),
		logger:      logger,
	}
}

// Use sets a Discord role ID required to run any subcommand. Empty string disables the check.
func (r *Router) Use(gmRoleID string) *Router {
	r.gmRoleID = gmRoleID
	return r
}

// Add registers a SubCommand under its Spec().Name. Returns r for chaining.
func (r *Router) Add(cmd SubCommand) *Router {
	r.subs[cmd.Spec().Name] = cmd
	return r
}

// ApplicationCommand builds the full *discordgo.ApplicationCommand from all registered subs.
func (r *Router) ApplicationCommand() *discordgo.ApplicationCommand {
	opts := make([]*discordgo.ApplicationCommandOption, 0, len(r.subs))
	for _, sub := range r.subs {
		opts = append(opts, sub.Spec())
	}
	return &discordgo.ApplicationCommand{
		Name:        r.name,
		Description: r.description,
		Options:     opts,
	}
}

// Handle is the InteractionCreate event handler registered with discordgo.
// It checks role access, then dispatches to the matching SubCommand.
func (r *Router) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if i.ApplicationCommandData().Name != r.name {
		return
	}

	ctx := context.Background()
	resp := &interactionContext{session: s, interaction: i.Interaction}

	if !r.hasAccess(i) {
		_ = resp.Send(ctx, "You need the GM role to use this command.", true)
		return
	}

	opts := i.ApplicationCommandData().Options
	if len(opts) == 0 {
		return
	}
	sub, ok := r.subs[opts[0].Name]
	if !ok {
		r.logger.Warn("unknown subcommand", "name", opts[0].Name)
		return
	}

	subOpts := newOptionMap(opts[0].Options)
	if err := sub.Handle(ctx, subOpts, resp); err != nil {
		r.logger.Error("subcommand error", "cmd", opts[0].Name, "err", err)
	}
}

func (r *Router) hasAccess(i *discordgo.InteractionCreate) bool {
	if r.gmRoleID == "" || i.Member == nil {
		return true
	}
	for _, role := range i.Member.Roles {
		if role == r.gmRoleID {
			return true
		}
	}
	return false
}
