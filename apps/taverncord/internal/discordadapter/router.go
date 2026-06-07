package discordadapter

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

type Router struct {
	name        string
	description string
	gmRoleID    string
	subs        map[string]SubCommand
	logger      *slog.Logger
}

func NewRouter(name, description string, logger *slog.Logger) *Router {
	return &Router{
		name:        name,
		description: description,
		subs:        make(map[string]SubCommand),
		logger:      logger,
	}
}

func (r *Router) Use(gmRoleID string) *Router {
	r.gmRoleID = gmRoleID
	return r
}

func (r *Router) Add(cmd SubCommand) *Router {
	r.subs[cmd.Spec().Name] = cmd
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
		_ = resp.Send(ctx, "You need the GM role to use this command.", command.Private)
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
