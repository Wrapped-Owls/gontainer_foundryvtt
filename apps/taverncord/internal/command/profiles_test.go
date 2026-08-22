package command

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"
)

func TestShowProfile_success(t *testing.T) {
	t.Parallel()

	client := &stubClient{profileInfo: ProfileInfo{
		Name: profAlice, DataPath: "/d", World: "w", HasAdminKey: true,
	}}
	resp := &stubResponder{}
	if err := makeCommands(client).ShowProfile(context.Background(), resp, profAlice); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{profAlice, "/d", "w", "yes"} {
		if !strings.Contains(resp.content, want) {
			t.Errorf("expected %q in response, got %q", want, resp.content)
		}
	}
	if !resp.ephemeral {
		t.Error("profile detail should be ephemeral")
	}
}

func TestShowProfile_notFound(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	client := &stubClient{}
	resp := &stubResponder{}
	if err := makeCommands(
		client,
	).EditProfile(context.Background(), resp, profAlice, ProfileInput{World: "w"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.lastName != profAlice || client.lastInput.World != "w" {
		t.Errorf("edit not forwarded: name=%q input=%+v", client.lastName, client.lastInput)
	}
}

const worldNew = "new"

func editProfile(t *testing.T, client stubClient, input ProfileInput) *stubResponder {
	t.Helper()

	resp := &stubResponder{}
	synctest.Test(t, func(t *testing.T) {
		if err := makeCommands(&client).EditProfile(
			context.Background(), resp, profAlice, input,
		); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	return resp
}

func activeProfile(world, version string) stubClient {
	return stubClient{
		status:      StatusData{Active: profAlice, Online: true},
		profileInfo: ProfileInfo{Name: profAlice, World: world, Version: version},
	}
}

func TestEditProfileAnnouncesARestart(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		client stubClient
		input  ProfileInput
	}{
		{
			name:   "changing the active profile's world",
			client: activeProfile("old", ""),
			input:  ProfileInput{World: worldNew},
		},
		{
			name:   "changing the active profile's version",
			client: activeProfile("", verFoundry13),
			input:  ProfileInput{Version: verFoundry14},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			resp := editProfile(t, testCase.client, testCase.input)
			if !strings.Contains(resp.content, "restarting") {
				t.Fatalf("reply %q should say the server is restarting", resp.content)
			}
			if resp.ephemeral {
				t.Error("a restart affects everyone, so the notice must not be ephemeral")
			}
		})
	}
}

func TestEditProfileStaysQuiet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		client stubClient
		input  ProfileInput
	}{
		{
			name:   "relabelling the active profile",
			client: activeProfile("old", ""),
			input:  ProfileInput{Label: "Master Alice"},
		},
		{
			name:   "setting the world it already has",
			client: activeProfile("same", ""),
			input:  ProfileInput{World: "same"},
		},
		{
			name: "editing an inactive profile",
			client: stubClient{
				status:      StatusData{Active: profBob, Online: true},
				profileInfo: ProfileInfo{Name: profAlice, World: "old"},
			},
			input: ProfileInput{World: worldNew},
		},
		{
			name: "an unreachable manager must not promise a restart",
			client: stubClient{
				statusErr:   errors.New("connection refused"),
				profileInfo: ProfileInfo{Name: profAlice, World: "old"},
			},
			input: ProfileInput{World: worldNew},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			resp := editProfile(t, testCase.client, testCase.input)
			if strings.Contains(resp.content, "restarting") {
				t.Fatalf("reply %q should not mention a restart", resp.content)
			}
		})
	}
}
