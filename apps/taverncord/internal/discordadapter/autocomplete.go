package discordadapter

import "github.com/bwmarrin/discordgo"

// respondChoices answers an autocomplete interaction. It answers on every path,
// including with no choices: Discord renders a missing response as "loading
// options failed" rather than as an empty list.
func respondChoices(
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
