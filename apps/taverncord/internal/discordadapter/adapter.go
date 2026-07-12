package discordadapter

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/config"
)

type Adapter struct {
	session    *discordgo.Session
	appID      string
	guildID    string
	router     *Router
	registered []*discordgo.ApplicationCommand
	logger     *slog.Logger
}

func New(cfg config.Config, router *Router, logger *slog.Logger) (*Adapter, error) {
	session, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return nil, fmt.Errorf("discord session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentsGuilds
	session.AddHandler(router.Handle)
	return &Adapter{
		session: session,
		appID:   cfg.Discord.ApplicationID,
		guildID: cfg.Discord.GuildID,
		router:  router,
		logger:  logger,
	}, nil
}

func (a *Adapter) Open() error {
	if err := a.session.Open(); err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}
	cmd, err := a.session.ApplicationCommandCreate(
		a.appID,
		a.guildID,
		a.router.ApplicationCommand(),
	)
	if err != nil {
		return fmt.Errorf("register commands: %w", err)
	}
	a.registered = append(a.registered, cmd)
	scope := "globally"
	if a.guildID != "" {
		scope = "for guild " + a.guildID
	}
	a.logger.Info("discord commands registered", "scope", scope)
	return nil
}

func (a *Adapter) SendMessage(channelID, content string) error {
	if _, err := a.session.ChannelMessageSend(channelID, content); err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

func (a *Adapter) Close() error {
	for _, cmd := range a.registered {
		if err := a.session.ApplicationCommandDelete(a.appID, a.guildID, cmd.ID); err != nil {
			a.logger.Warn("failed to delete command", "cmd", cmd.Name, "err", err)
		}
	}
	return a.session.Close()
}
