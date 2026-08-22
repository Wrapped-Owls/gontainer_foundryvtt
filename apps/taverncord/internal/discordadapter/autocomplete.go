package discordadapter

import "github.com/bwmarrin/discordgo"

func respondChoices( // always answers so Discord doesn't show "loading options failed"
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	values []string,
) error {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(values))
	for _, value := range values {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  value,
			Value: value,
		})
	}
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{Choices: choices},
	})
}
