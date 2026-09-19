package discordadapter

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

type interactionContext struct {
	session     *discordgo.Session
	interaction *discordgo.Interaction
}

func (c *interactionContext) Send(
	_ context.Context,
	content string,
	visibility command.Visibility,
) error {
	var flags discordgo.MessageFlags
	if visibility == command.Private {
		flags = discordgo.MessageFlagsEphemeral
	}
	return c.session.InteractionRespond(c.interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   flags,
		},
	})
}

func (c *interactionContext) Edit(_ context.Context, content string) error {
	_, err := c.session.InteractionResponseEdit(c.interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	return err
}
