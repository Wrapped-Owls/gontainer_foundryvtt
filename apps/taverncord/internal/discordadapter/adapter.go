package discordadapter

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/config"
)

// Adapter manages the discordgo session lifecycle and command registration.
type Adapter struct {
	session    *discordgo.Session
	appID      string
	guildID    string
	router     *Router
	registered []*discordgo.ApplicationCommand
	logger     *slog.Logger
}

// New creates an Adapter from the provided config and router.
func New(cfg config.Config, router *Router, logger *slog.Logger) (*Adapter, error) {
	session, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return nil, fmt.Errorf("discord session: %w", err)
	}
	// IntentsGuilds is the minimum required for the bot to appear online and
	// receive application command interactions within guild channels.
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

// Open connects to Discord and registers the slash commands.
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

// Close deletes registered commands and closes the Discord session.
func (a *Adapter) Close() error {
	for _, cmd := range a.registered {
		if err := a.session.ApplicationCommandDelete(a.appID, a.guildID, cmd.ID); err != nil {
			a.logger.Warn("failed to delete command", "cmd", cmd.Name, "err", err)
		}
	}
	return a.session.Close()
}
