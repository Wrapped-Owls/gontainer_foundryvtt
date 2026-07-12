package command

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

type stubClient struct {
	profiles     ProfilesData
	status       StatusData
	switchErr    error
	listErr      error
	statusErr    error
	gotInterrupt Interrupt
}

func (s *stubClient) ListProfiles(_ context.Context) (ProfilesData, error) {
	return s.profiles, s.listErr
}

func (s *stubClient) Switch(_ context.Context, _ string, interrupt Interrupt) error {
	s.gotInterrupt = interrupt
	return s.switchErr
}

func (s *stubClient) Status(_ context.Context) (StatusData, error) {
	return s.status, s.statusErr
}

type stubResponder struct {
	content    string
	visibility Visibility
	edited     string
}

func (r *stubResponder) Send(_ context.Context, content string, visibility Visibility) error {
	r.content = content
	r.visibility = visibility
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
	if resp.visibility != Private {
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
	if err := makeCommands(
		&stubClient{},
	).Switch(context.Background(), resp, "bob", InterruptWhenIdle); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.visibility != Public {
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
	if err := makeCommands(
		client,
	).Switch(context.Background(), resp, "nobody", InterruptWhenIdle); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.edited, "❌") {
		t.Errorf("expected failure marker in edited response, got %q", resp.edited)
	}
	if !strings.Contains(resp.edited, "unknown profile") {
		t.Errorf("expected error detail in edited response, got %q", resp.edited)
	}
}

func TestSwitch_passesInterrupt(t *testing.T) {
	client := &stubClient{}
	resp := &stubResponder{}
	if err := makeCommands(
		client,
	).Switch(context.Background(), resp, "bob", InterruptAlways); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.gotInterrupt != InterruptAlways {
		t.Errorf("interrupt forwarded as %q, want %q", client.gotInterrupt, InterruptAlways)
	}
}

func TestStatus_offline(t *testing.T) {
	client := &stubClient{status: StatusData{Active: "alice", Version: "14.0.0", Online: false}}
	resp := &stubResponder{}
	if err := makeCommands(client).Status(context.Background(), resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.content, "alice") || !strings.Contains(resp.content, "14.0.0") {
		t.Errorf("expected active+version in response, got %q", resp.content)
	}
	if !strings.Contains(resp.content, "offline") {
		t.Errorf("expected offline marker, got %q", resp.content)
	}
}

func TestStatus_online(t *testing.T) {
	client := &stubClient{status: StatusData{
		Active:        "alice",
		Version:       "13.351",
		Online:        true,
		WorldActive:   true,
		World:         "my-world",
		System:        "projectfu",
		SystemVersion: "4.16.1",
		Users:         3,
		UptimeMS:      6230770,
	}}
	resp := &stubResponder{}
	if err := makeCommands(client).Status(context.Background(), resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"online", "my-world", "projectfu", "3"} {
		if !strings.Contains(resp.content, want) {
			t.Errorf("expected %q in response, got %q", want, resp.content)
		}
	}
}
