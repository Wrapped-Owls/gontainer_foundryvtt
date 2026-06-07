package discordadapter

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// interactionContext implements command.Responder for a single Discord interaction.
type interactionContext struct {
	session     *discordgo.Session
	interaction *discordgo.Interaction
}

// Send replies to the interaction. When ephemeral is true the reply is only
// visible to the invoking user.
func (c *interactionContext) Send(_ context.Context, content string, ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
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

// Edit replaces the content of the initial response already sent by Send.
func (c *interactionContext) Edit(_ context.Context, content string) error {
	_, err := c.session.InteractionResponseEdit(c.interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	return err
}
