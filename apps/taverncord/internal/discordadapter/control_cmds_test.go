package discordadapter

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

func TestInterruptOf(t *testing.T) {
	t.Parallel()

	forceSetTo := func(isForced bool) OptionMap {
		return OptionMap{optionForce: {
			Name:  optionForce,
			Type:  discordgo.ApplicationCommandOptionBoolean,
			Value: isForced,
		}}
	}
	testCases := []struct {
		name string
		opts OptionMap
		want command.Interrupt
	}{
		{name: "force omitted", opts: OptionMap{}, want: command.InterruptWhenIdle},
		{name: "force false", opts: forceSetTo(false), want: command.InterruptWhenIdle},
		{name: "force true", opts: forceSetTo(true), want: command.InterruptAlways},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := interruptOf(testCase.opts); got != testCase.want {
				t.Fatalf("interruptOf = %q, want %q", got, testCase.want)
			}
		})
	}
}
