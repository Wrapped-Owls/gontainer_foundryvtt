package discordadapter

import (
	"context"
	"log/slog"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

// stubSubCmd is a test SubCommand that records calls.
type stubSubCmd struct {
	name   string
	called bool
}

func (s *stubSubCmd) Spec() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        s.name,
		Description: "stub",
	}
}

func (s *stubSubCmd) Handle(_ context.Context, _ OptionMap, r command.Responder) error {
	s.called = true
	return r.Send(context.Background(), "ok", true)
}

// stubResponder records the last Send and Edit calls.
type stubResponder struct {
	content   string
	ephemeral bool
	edited    string
}

func (r *stubResponder) Send(_ context.Context, content string, ephemeral bool) error {
	r.content = content
	r.ephemeral = ephemeral
	return nil
}

func (r *stubResponder) Edit(_ context.Context, content string) error {
	r.edited = content
	return nil
}

func TestRouter_ApplicationCommand_containsSubcommands(t *testing.T) {
	router := NewRouter("foundry", "desc", slog.Default()).
		Add(&stubSubCmd{name: "list"}).
		Add(&stubSubCmd{name: "status"})

	cmd := router.ApplicationCommand()
	if cmd.Name != "foundry" {
		t.Errorf("expected name foundry, got %q", cmd.Name)
	}
	if len(cmd.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(cmd.Options))
	}
}

func TestRouter_hasAccess_noRole(t *testing.T) {
	router := NewRouter("foundry", "desc", slog.Default())
	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		Member: &discordgo.Member{Roles: []string{"other"}},
	}}
	if !router.hasAccess(i) {
		t.Error("expected access when gmRoleID is empty")
	}
}

func TestRouter_hasAccess_roleMatch(t *testing.T) {
	router := NewRouter("foundry", "desc", slog.Default()).Use("gm-role-id")
	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		Member: &discordgo.Member{Roles: []string{"other", "gm-role-id"}},
	}}
	if !router.hasAccess(i) {
		t.Error("expected access for member with GM role")
	}
}

func TestRouter_hasAccess_roleMismatch(t *testing.T) {
	router := NewRouter("foundry", "desc", slog.Default()).Use("gm-role-id")
	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		Member: &discordgo.Member{Roles: []string{"other"}},
	}}
	if router.hasAccess(i) {
		t.Error("expected no access for member without GM role")
	}
}
