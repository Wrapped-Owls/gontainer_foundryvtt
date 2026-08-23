package discordadapter

import (
	"context"
	"log/slog"
	"slices"

	"github.com/bwmarrin/discordgo"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

type Router struct {
	ctx         context.Context
	name        string
	description string
	gmRoleID    string
	subs        map[string]subCommand
	logger      *slog.Logger
}

func NewRouter(ctx context.Context, name, description string, logger *slog.Logger) *Router {
	return &Router{
		ctx:         ctx,
		name:        name,
		description: description,
		subs:        make(map[string]subCommand),
		logger:      logger,
	}
}

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
		Name:         r.name,
		Description:  r.description,
		Options:      opts,
		DMPermission: new(false),
	}
}

func (r *Router) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer func() { // discordgo runs handlers with no recover: a panic here ends the process
		if v := recover(); v != nil {
			r.logger.Error("interaction handler panicked", "panic", v)
		}
	}()

	switch i.Type { // ApplicationCommandData panics on any other interaction type
	case discordgo.InteractionApplicationCommand:
		if invoked := i.ApplicationCommandData(); invoked.Name == r.name {
			r.handleCommand(s, i, invoked)
		}
	case discordgo.InteractionApplicationCommandAutocomplete:
		if invoked := i.ApplicationCommandData(); invoked.Name == r.name {
			if err := respondChoices(s, i, r.autocompleteChoices(invoked, i.Member)); err != nil {
				r.logger.Warn("autocomplete response failed", "err", err)
			}
		}
	}
}

func (r *Router) handleCommand(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	invoked discordgo.ApplicationCommandInteractionData,
) {
	resp := &interactionContext{session: s, interaction: i.Interaction}

	if !r.hasAccess(i.Member) {
		_ = resp.Send(r.ctx, "You need the GM role to use this command.", command.Private)
		return
	}

	inv, isSub := parseInvocation(invoked)
	if !isSub {
		_ = resp.Send(r.ctx, "No subcommand given. Try `/foundry status`.", command.Private)
		return
	}
	sub, isKnown := r.subs[inv.Name]
	if !isKnown {
		r.logger.Warn("unknown subcommand", "name", inv.Name)
		_ = resp.Send(r.ctx, "That subcommand is not available.", command.Private)
		return
	}

	if err := sub.Handle(r.ctx, newOptionMap(inv.Options), resp); err != nil {
		r.logger.Error("subcommand error", "cmd", inv.Name, "err", err)
	}
}

func (r *Router) autocompleteChoices(
	invoked discordgo.ApplicationCommandInteractionData,
	member *discordgo.Member,
) []string {
	if !r.hasAccess(member) {
		return nil
	}
	inv, isSub := parseInvocation(invoked)
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
	if r.gmRoleID == "" {
		return true
	}
	if member == nil {
		return false
	}
	return slices.Contains(member.Roles, r.gmRoleID)
}
