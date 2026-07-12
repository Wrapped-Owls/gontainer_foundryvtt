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

func TestCreateProfile_forwardsInput(t *testing.T) {
	client := &stubClient{}
	resp := &stubResponder{}
	in := ProfileInput{Name: "bob", DataPath: "/d/bob", World: "w"}
	if err := makeCommands(client).CreateProfile(context.Background(), resp, in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.lastInput.Name != "bob" || client.lastInput.DataPath != "/d/bob" {
		t.Errorf("input not forwarded: %+v", client.lastInput)
	}
	if !strings.Contains(resp.content, "✅") {
		t.Errorf("expected success marker, got %q", resp.content)
	}
}

func TestCreateProfile_failureRelaysError(t *testing.T) {
	client := &stubClient{createErr: errors.New("profile: already exists")}
	resp := &stubResponder{}
	if err := makeCommands(
		client,
	).CreateProfile(context.Background(), resp, ProfileInput{Name: "bob"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.content, "already exists") {
		t.Errorf("expected relayed error, got %q", resp.content)
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

func TestDeleteProfile_failureRelaysError(t *testing.T) {
	client := &stubClient{deleteErr: errors.New("cannot delete the active profile")}
	resp := &stubResponder{}
	if err := makeCommands(client).DeleteProfile(context.Background(), resp, "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.content, "active profile") {
		t.Errorf("expected relayed error, got %q", resp.content)
	}
}
