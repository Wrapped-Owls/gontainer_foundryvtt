package discordadapter

import (
	"context"
	"log/slog"
	"slices"

	"github.com/bwmarrin/discordgo"
)

// Router maps subcommand names to SubCommand implementations and dispatches
// Discord interactions, analogous to http.ServeMux.
type Router struct {
	ctx         context.Context
	name        string
	description string
	gmRoleID    string
	subs        map[string]subCommand
	logger      *slog.Logger
}

// NewRouter creates a Router for a top-level slash command. ctx bounds every
// handler: a switch or restart blocks for as long as Foundry takes to come back,
// so without it a shutdown would leave those waits running.
func NewRouter(ctx context.Context, name, description string, logger *slog.Logger) *Router {
	return &Router{
		ctx:         ctx,
		name:        name,
		description: description,
		subs:        make(map[string]subCommand),
		logger:      logger,
	}
}

// Use sets a Discord role ID required to run any subcommand. Empty string disables the check.
func (r *Router) Use(gmRoleID string) *Router {
	r.gmRoleID = gmRoleID
	return r
}

func (r *Router) Add(subs ...subCommand) *Router {
	for _, sub := range subs {
		r.subs[sub.Spec().Name] = sub
	}
	return r
}

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

// Handle is the InteractionCreate event handler registered with discordgo, which
// runs it in its own goroutine with no recover: an unhandled panic here ends the
// process. The type check guards the one panic discordgo itself raises, and the
// recover catches whatever a subcommand does.
func (r *Router) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer func() {
		if v := recover(); v != nil {
			r.logger.Error("interaction handler panicked", "panic", v)
		}
	}()

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if data := i.ApplicationCommandData(); data.Name == r.name {
			r.handleCommand(s, i, data)
		}
	case discordgo.InteractionApplicationCommandAutocomplete:
		if data := i.ApplicationCommandData(); data.Name == r.name {
			if err := respondChoices(s, i, r.autocompleteChoices(data, i.Member)); err != nil {
				r.logger.Warn("autocomplete response failed", "err", err)
			}
		}
	}
}

// handleCommand answers on every path: an unanswered interaction shows in Discord
// as a bare "interaction failed".
func (r *Router) handleCommand(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	data discordgo.ApplicationCommandInteractionData,
) {
	resp := &interactionContext{session: s, interaction: i.Interaction}

	if !r.hasAccess(i.Member) {
		_ = resp.Send(r.ctx, "You need the GM role to use this command.", true)
		return
	}

	inv, isSub := parseInvocation(data)
	if !isSub {
		_ = resp.Send(r.ctx, "No subcommand given. Try `/foundry status`.", true)
		return
	}
	sub, isKnown := r.subs[inv.Name]
	if !isKnown {
		r.logger.Warn("unknown subcommand", "name", inv.Name)
		_ = resp.Send(r.ctx, "That subcommand is not available.", true)
		return
	}

	if err := sub.Handle(r.ctx, newOptionMap(inv.Options), resp); err != nil {
		r.logger.Error("subcommand error", "cmd", inv.Name, "err", err)
	}
}

// autocompleteChoices answers Discord's as-you-type request. The role gate applies
// here too: profile and version names are not for the whole channel to enumerate.
func (r *Router) autocompleteChoices(
	data discordgo.ApplicationCommandInteractionData,
	member *discordgo.Member,
) []string {
	if !r.hasAccess(member) {
		return nil
	}
	inv, isSub := parseInvocation(data)
	if !isSub {
		return nil
	}
	sub, isKnown := r.subs[inv.Name]
	if !isKnown {
		return nil
	}
	focused, typed := focusedOption(inv.Options)
	if focused == "" {
		return nil
	}
	return sub.Autocomplete(r.ctx, focused, typed)
}

func (r *Router) hasAccess(member *discordgo.Member) bool {
	if r.gmRoleID == "" || member == nil {
		return true
	}
	return slices.Contains(member.Roles, r.gmRoleID)
}
