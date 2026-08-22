package discordadapter

import (
	"context"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

const (
	roleGM    = "gm-role-id"
	roleOther = "other"
	profAlice = "alice"
	profBob   = "bob"
)

func discardLogger() *slog.Logger { return slog.New(slog.DiscardHandler) }

type stubFoundryClient struct {
	profiles command.ProfilesData
	versions command.VersionsData
	status   command.StatusData
}

var _ command.FoundryClient = (*stubFoundryClient)(nil)

func (s *stubFoundryClient) ListProfiles(context.Context) (command.ProfilesData, error) {
	return s.profiles, nil
}
func (*stubFoundryClient) Switch(context.Context, string, bool) error { return nil }
func (*stubFoundryClient) Restart(context.Context, bool) error        { return nil }
func (s *stubFoundryClient) Status(context.Context) (command.StatusData, error) {
	return s.status, nil
}

func (s *stubFoundryClient) Versions(context.Context) (command.VersionsData, error) {
	return s.versions, nil
}
func (*stubFoundryClient) Download(context.Context, string, string) error { return nil }
func (*stubFoundryClient) GetProfile(_ context.Context, name string) (command.ProfileInfo, error) {
	return command.ProfileInfo{Name: name}, nil
}

func (*stubFoundryClient) UpdateProfile(context.Context, string, command.ProfileInput) error {
	return nil
}

func (*stubFoundryClient) Logs(context.Context, int) (command.LogsData, error) {
	return command.LogsData{}, nil
}

func (*stubFoundryClient) Events(context.Context, int) (command.EventsData, error) {
	return command.EventsData{}, nil
}
