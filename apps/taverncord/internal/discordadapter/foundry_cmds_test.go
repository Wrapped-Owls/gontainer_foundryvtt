package discordadapter

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

// stubFoundryClient implements command.FoundryClient for testing.
type stubFoundryClient struct {
	profiles  command.ProfilesData
	status    command.StatusData
	switchErr error
}

func (s *stubFoundryClient) ListProfiles(_ context.Context) (command.ProfilesData, error) {
	return s.profiles, nil
}
func (s *stubFoundryClient) Switch(_ context.Context, _ string, _ bool) error { return s.switchErr }
func (s *stubFoundryClient) Status(_ context.Context) (command.StatusData, error) {
	return s.status, nil
}
func (s *stubFoundryClient) Versions(_ context.Context) (command.VersionsData, error) {
	return command.VersionsData{}, nil
}
func (s *stubFoundryClient) Download(_ context.Context, _, _ string) error { return nil }
func (s *stubFoundryClient) GetProfile(_ context.Context, _ string) (command.ProfileInfo, error) {
	return command.ProfileInfo{}, nil
}
func (s *stubFoundryClient) CreateProfile(_ context.Context, _ command.ProfileInput) error {
	return nil
}
func (s *stubFoundryClient) UpdateProfile(_ context.Context, _ string, _ command.ProfileInput) error {
	return nil
}
func (s *stubFoundryClient) DeleteProfile(_ context.Context, _ string) error { return nil }
func (s *stubFoundryClient) Logs(_ context.Context, _ int) (command.LogsData, error) {
	return command.LogsData{}, nil
}
func (s *stubFoundryClient) Events(_ context.Context, _ int) (command.EventsData, error) {
	return command.EventsData{}, nil
}

func makeProfileCmds(client command.FoundryClient) *command.ProfileCommands {
	return command.New(client, slog.Default())
}

func TestListCmd_spec(t *testing.T) {
	spec := ListCmd(makeProfileCmds(&stubFoundryClient{})).Spec()
	if spec.Name != "list" {
		t.Errorf("expected name list, got %q", spec.Name)
	}
}

func TestSwitchCmd_spec(t *testing.T) {
	spec := SwitchCmd(makeProfileCmds(&stubFoundryClient{})).Spec()
	if spec.Name != "switch" {
		t.Errorf("expected name switch, got %q", spec.Name)
	}
	if len(spec.Options) == 0 {
		t.Error("switch spec should declare the name option")
	}
}

func TestStatusCmd_spec(t *testing.T) {
	spec := StatusCmd(makeProfileCmds(&stubFoundryClient{})).Spec()
	if spec.Name != "status" {
		t.Errorf("expected name status, got %q", spec.Name)
	}
}

func TestSwitchCmd_failure_editsMessage(t *testing.T) {
	client := &stubFoundryClient{switchErr: errors.New("unknown profile")}
	resp := &stubResponder{}
	cmd := SwitchCmd(makeProfileCmds(client))
	if err := cmd.Handle(context.Background(), OptionMap{}, resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.edited, "❌") {
		t.Errorf("expected failure marker in edited response, got %q", resp.edited)
	}
}

func TestSwitchCmd_success_editsMessage(t *testing.T) {
	resp := &stubResponder{}
	cmd := SwitchCmd(makeProfileCmds(&stubFoundryClient{}))
	if err := cmd.Handle(context.Background(), OptionMap{}, resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.edited, "✅") {
		t.Errorf("expected success marker in edited response, got %q", resp.edited)
	}
}
