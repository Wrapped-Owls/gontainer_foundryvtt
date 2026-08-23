package discordadapter

import (
	"context"
	"log/slog"
	"slices"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

func TestApplicationCommandContainsSubcommands(t *testing.T) {
	t.Parallel()

	router := NewRouter(context.Background(), "foundry", "desc", slog.Default()).
		Add(subCommand{name: subList, description: "stub"}).
		Add(subCommand{name: subStatus, description: "stub"})

	cmd := router.ApplicationCommand()
	if cmd.Name != "foundry" {
		t.Errorf("expected name foundry, got %q", cmd.Name)
	}
	if len(cmd.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(cmd.Options))
	}
	if cmd.DMPermission == nil || *cmd.DMPermission {
		t.Errorf("expected DMPermission false, got %v", cmd.DMPermission)
	}
}

func TestRouterHasAccess(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		gmRoleID string
		member   *discordgo.Member
		want     bool
	}{
		{
			name:   "an unset role gate lets everyone through",
			member: &discordgo.Member{Roles: []string{roleOther}},
			want:   true,
		},
		{
			name:     "an unset role gate lets even a nil member through",
			gmRoleID: "",
			member:   nil,
			want:     true,
		},
		{
			name:     "a member carrying the role is allowed",
			gmRoleID: roleGM,
			member:   &discordgo.Member{Roles: []string{roleOther, roleGM}},
			want:     true,
		},
		{
			name:     "a member without the role is denied",
			gmRoleID: roleGM,
			member:   &discordgo.Member{Roles: []string{roleOther}},
		},
		{
			name:     "a nil member is denied when a role is configured: DMs carry no member",
			gmRoleID: roleGM,
			member:   nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			router := NewRouter(context.Background(), "foundry", "desc", discardLogger()).
				Use(testCase.gmRoleID)
			if got := router.hasAccess(testCase.member); got != testCase.want {
				t.Fatalf("hasAccess() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestAutocompleteChoicesRespectsTheRoleGate(t *testing.T) {
	t.Parallel()

	client := &stubFoundryClient{profiles: command.ProfilesData{
		Profiles: []profile.Profile{{Name: profAlice}, {Name: profBob}},
	}}
	router := NewRouter(context.Background(), "foundry", "desc", discardLogger()).
		Use(roleGM).
		Add(ControlCommands(command.New(client, discardLogger()))...)

	testCases := []struct {
		name    string
		member  *discordgo.Member
		subName string
		focused string
		want    []string
	}{
		{
			name:    "a member without the GM role is offered nothing",
			member:  &discordgo.Member{Roles: []string{"everyone"}},
			subName: subSwitch,
			focused: optionName,
		},
		{
			name:    "a member with the GM role is offered the profiles",
			member:  &discordgo.Member{Roles: []string{roleGM}},
			subName: subSwitch,
			focused: optionName,
			want:    []string{profAlice, profBob},
		},
		{
			name:    "an unknown subcommand is offered nothing",
			member:  &discordgo.Member{Roles: []string{roleGM}},
			subName: "nonsense",
			focused: optionName,
		},
		{
			name:    "an option with no suggester is offered nothing",
			member:  &discordgo.Member{Roles: []string{roleGM}},
			subName: subSwitch,
			focused: optionForce,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			invoked := subData(testCase.subName, subOption(testCase.focused, true, ""))
			got := router.autocompleteChoices(invoked, testCase.member)
			if !slices.Equal(got, testCase.want) {
				t.Fatalf("got %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestHandleSurvivesAPanickingInteraction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		interaction discordgo.Interaction
	}{
		{
			name: "a subcommand that panics does not end the process",
			interaction: discordgo.Interaction{
				Type: discordgo.InteractionApplicationCommand,
				Data: discordgo.ApplicationCommandInteractionData{Name: "foundry"},
			},
		},
		{
			name: "an autocomplete that panics does not end the process",
			interaction: discordgo.Interaction{
				Type: discordgo.InteractionApplicationCommandAutocomplete,
				Data: discordgo.ApplicationCommandInteractionData{Name: "foundry"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			router := NewRouter(t.Context(), "foundry", "desc", slog.Default()).
				Add(subCommand{name: subStatus, description: "stub"})

			router.Handle(nil, &discordgo.InteractionCreate{Interaction: &testCase.interaction})
		})
	}
}
