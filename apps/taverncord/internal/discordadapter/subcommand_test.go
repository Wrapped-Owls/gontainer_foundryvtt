package discordadapter

import (
	"context"
	"slices"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

func subOption(
	name string,
	focused bool,
	value string,
) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name:    name,
		Type:    discordgo.ApplicationCommandOptionString,
		Value:   value,
		Focused: focused,
	}
}

func subData(
	name string,
	opts ...*discordgo.ApplicationCommandInteractionDataOption,
) discordgo.ApplicationCommandInteractionData {
	return discordgo.ApplicationCommandInteractionData{
		Options: []*discordgo.ApplicationCommandInteractionDataOption{{
			Name:    name,
			Type:    discordgo.ApplicationCommandOptionSubCommand,
			Options: opts,
		}},
	}
}

func TestParseInvocationFindsTheSubcommand(t *testing.T) {
	t.Parallel()

	inv, isSub := parseInvocation(subData(subSwitch, subOption(optionName, false, profAlice)))
	if !isSub {
		t.Fatal("a plain subcommand should parse")
	}
	if inv.Name != subSwitch || len(inv.Options) != 1 {
		t.Fatalf("got %q with %d options, want %q with 1", inv.Name, len(inv.Options), subSwitch)
	}
}

func TestParseInvocationRejectsWhatIsNotASubcommand(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		data discordgo.ApplicationCommandInteractionData
	}{
		{
			name: "no options at all",
			data: discordgo.ApplicationCommandInteractionData{},
		},
		{
			name: "a bare top-level option",
			data: discordgo.ApplicationCommandInteractionData{
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					subOption(optionName, false, profAlice),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if _, isSub := parseInvocation(testCase.data); isSub {
				t.Fatal("expected this not to parse as a subcommand invocation")
			}
		})
	}
}

func TestFocusedOption(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		opts      []*discordgo.ApplicationCommandInteractionDataOption
		wantName  string
		wantTyped string
	}{
		{name: "no options means nothing is focused"},
		{
			name: "the focused option and its partial value are reported",
			opts: []*discordgo.ApplicationCommandInteractionDataOption{
				subOption(optionLabel, false, "x"),
				subOption(optionName, true, "ali"),
			},
			wantName:  optionName,
			wantTyped: "ali",
		},
		{
			name: "an unfocused set reports nothing",
			opts: []*discordgo.ApplicationCommandInteractionDataOption{
				subOption(optionLabel, false, "x"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotName, gotTyped := focusedOption(testCase.opts)
			if gotName != testCase.wantName || gotTyped != testCase.wantTyped {
				t.Fatalf("got (%q, %q), want (%q, %q)",
					gotName, gotTyped, testCase.wantName, testCase.wantTyped)
			}
		})
	}
}

func TestSubCommandAutocompleteAnswersOnlyForItsOwnOptions(t *testing.T) {
	t.Parallel()

	sub := subCommand{
		name: subDownload,
		suggest: map[string]suggester{
			optionVersion: func(_ context.Context, typed string) []string {
				return []string{"got:" + typed}
			},
		},
	}

	testCases := []struct {
		name    string
		focused string
		want    []string
	}{
		{
			name:    "a registered option is completed",
			focused: optionVersion,
			want:    []string{"got:14"},
		},
		{
			name:    "an option with no suggester completes nothing",
			focused: optionURL,
		},
		{
			name:    "an unknown option completes nothing",
			focused: "nonsense",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := sub.Autocomplete(context.Background(), testCase.focused, "14")
			if !slices.Equal(got, testCase.want) {
				t.Fatalf("got %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestEverySubCommandSpecIsWellFormed(t *testing.T) {
	t.Parallel()

	cmds := command.New(&stubFoundryClient{}, discardLogger())
	subs := append(ReadCommands(cmds), ControlCommands(cmds)...)

	seen := make(map[string]bool, len(subs))
	for _, sub := range subs {
		spec := sub.Spec()
		t.Run(spec.Name, func(t *testing.T) {
			if spec.Type != discordgo.ApplicationCommandOptionSubCommand {
				t.Errorf("type = %v, want SubCommand", spec.Type)
			}
			if spec.Description == "" {
				t.Error("a subcommand with no description is rejected by Discord")
			}
			if seen[spec.Name] {
				t.Errorf("duplicate subcommand name %q", spec.Name)
			}
			seen[spec.Name] = true

			// An option that advertises Autocomplete without a registered
			// suggester leaves Discord completing against nothing, forever.
			for _, opt := range spec.Options {
				if _, hasSuggester := sub.suggest[opt.Name]; opt.Autocomplete != hasSuggester {
					t.Errorf(
						"option %q: Autocomplete=%v but suggester registered=%v",
						opt.Name, opt.Autocomplete, hasSuggester,
					)
				}
			}
		})
	}
}
