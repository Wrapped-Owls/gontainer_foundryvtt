package command

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

// stubClient implements FoundryClient for testing.
type stubClient struct {
	profiles  ProfilesData
	status    StatusData
	switchErr error
	listErr   error
	statusErr error
}

func (s *stubClient) ListProfiles(_ context.Context) (ProfilesData, error) {
	return s.profiles, s.listErr
}
func (s *stubClient) Switch(_ context.Context, _ string) error { return s.switchErr }
func (s *stubClient) Status(_ context.Context) (StatusData, error) {
	return s.status, s.statusErr
}

// stubResponder captures Send and Edit calls for assertion.
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

func makeCommands(client FoundryClient) *ProfileCommands {
	return New(client, slog.Default())
}

func TestList_marksActiveProfile(t *testing.T) {
	client := &stubClient{profiles: ProfilesData{
		Active: "alice",
		Profiles: []profile.Profile{
			{Name: "alice", Label: "Alice"},
			{Name: "bob", Label: "Bob"},
		},
	}}
	resp := &stubResponder{}
	if err := makeCommands(client).List(context.Background(), resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.content, "▶") {
		t.Error("expected active marker ▶ in response")
	}
	if !strings.Contains(resp.content, "○") {
		t.Error("expected inactive marker ○ in response")
	}
	if !resp.ephemeral {
		t.Error("list response should be ephemeral")
	}
}

func TestList_clientError(t *testing.T) {
	client := &stubClient{listErr: errors.New("connection refused")}
	resp := &stubResponder{}
	if err := makeCommands(client).List(context.Background(), resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.content, "Failed") {
		t.Errorf("expected failure message, got %q", resp.content)
	}
}

func TestSwitch_success_editsMessage(t *testing.T) {
	resp := &stubResponder{}
	if err := makeCommands(&stubClient{}).Switch(context.Background(), resp, "bob"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ephemeral {
		t.Error("initial switch acknowledgement should not be ephemeral")
	}
	if !strings.Contains(resp.edited, "bob") {
		t.Errorf("expected profile name in edited response, got %q", resp.edited)
	}
	if !strings.Contains(resp.edited, "✅") {
		t.Errorf("expected success marker in edited response, got %q", resp.edited)
	}
}

func TestSwitch_failure_editsMessage(t *testing.T) {
	client := &stubClient{switchErr: errors.New("unknown profile")}
	resp := &stubResponder{}
	if err := makeCommands(client).Switch(context.Background(), resp, "nobody"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.edited, "❌") {
		t.Errorf("expected failure marker in edited response, got %q", resp.edited)
	}
	if !strings.Contains(resp.edited, "unknown profile") {
		t.Errorf("expected error detail in edited response, got %q", resp.edited)
	}
}

func TestStatus_success(t *testing.T) {
	client := &stubClient{status: StatusData{Active: "alice", Version: "14.0.0"}}
	resp := &stubResponder{}
	if err := makeCommands(client).Status(context.Background(), resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.content, "alice") || !strings.Contains(resp.content, "14.0.0") {
		t.Errorf("expected active+version in response, got %q", resp.content)
	}
}
