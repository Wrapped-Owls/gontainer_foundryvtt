package command

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestShowProfile_success(t *testing.T) {
	client := &stubClient{profileInfo: ProfileInfo{
		Name: "alice", DataPath: "/d", World: "w", HasAdminKey: true,
	}}
	resp := &stubResponder{}
	if err := makeCommands(client).ShowProfile(context.Background(), resp, "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"alice", "/d", "w", "yes"} {
		if !strings.Contains(resp.content, want) {
			t.Errorf("expected %q in response, got %q", want, resp.content)
		}
	}
	if !resp.ephemeral {
		t.Error("profile detail should be ephemeral")
	}
}

func TestShowProfile_notFound(t *testing.T) {
	client := &stubClient{profileErr: errors.New("profile not found")}
	resp := &stubResponder{}
	if err := makeCommands(client).ShowProfile(context.Background(), resp, "ghost"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.content, "❌") {
		t.Errorf("expected failure marker, got %q", resp.content)
	}
}

func TestEditProfile_forwardsName(t *testing.T) {
	client := &stubClient{}
	resp := &stubResponder{}
	if err := makeCommands(
		client,
	).EditProfile(context.Background(), resp, "alice", ProfileInput{World: "w"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.lastName != "alice" || client.lastInput.World != "w" {
		t.Errorf("edit not forwarded: name=%q input=%+v", client.lastName, client.lastInput)
	}
}
