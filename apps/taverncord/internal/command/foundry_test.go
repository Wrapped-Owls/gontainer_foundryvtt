package command

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

const (
	profAlice      = "alice"
	profAliceLabel = "Alice"
	profBob        = "bob"

	worldSharn   = "sharn"
	worldEberron = "eberron"
	verFoundry13 = "13.351.0"
	verFoundry14 = "14.361.0"
)

type stubClient struct {
	profiles    ProfilesData
	status      StatusData
	versions    VersionsData
	profileInfo ProfileInfo
	logs        LogsData
	events      EventsData
	statusSeq   []StatusData
	statusCalls int
	switchErr   error
	restartErr  error
	listErr     error
	statusErr   error
	versionsErr error
	downloadErr error
	profileErr  error
	updateErr   error
	logsErr     error
	eventsErr   error
	gotForce    bool
	restarts    int
	gotVersion  string
	gotURL      string
	gotTail     int
	lastName    string
	lastInput   ProfileInput
}

func (s *stubClient) ListProfiles(_ context.Context) (ProfilesData, error) {
	return s.profiles, s.listErr
}

func (s *stubClient) Switch(_ context.Context, _ string, force bool) error {
	s.gotForce = force
	return s.switchErr
}

func (s *stubClient) Restart(_ context.Context, force bool) error {
	s.restarts++
	s.gotForce = force
	return s.restartErr
}

func (s *stubClient) Status(_ context.Context) (StatusData, error) {
	if len(s.statusSeq) == 0 {
		return s.status, s.statusErr
	}
	idx := min(s.statusCalls, len(s.statusSeq)-1)
	s.statusCalls++
	return s.statusSeq[idx], s.statusErr
}

func (s *stubClient) Versions(_ context.Context) (VersionsData, error) {
	return s.versions, s.versionsErr
}

func (s *stubClient) Download(_ context.Context, version, url string) error {
	s.gotVersion, s.gotURL = version, url
	return s.downloadErr
}

func (s *stubClient) GetProfile(_ context.Context, _ string) (ProfileInfo, error) {
	return s.profileInfo, s.profileErr
}

func (s *stubClient) UpdateProfile(_ context.Context, name string, p ProfileInput) error {
	s.lastName, s.lastInput = name, p
	return s.updateErr
}

func (s *stubClient) Logs(_ context.Context, tail int) (LogsData, error) {
	s.gotTail = tail
	return s.logs, s.logsErr
}

func (s *stubClient) Events(_ context.Context, _ int) (EventsData, error) {
	return s.events, s.eventsErr
}

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
	t.Parallel()

	client := &stubClient{profiles: ProfilesData{
		Active: profAlice,
		Profiles: []profile.Profile{
			{Name: profAlice, Label: profAliceLabel},
			{Name: profBob, Label: "Bob"},
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

func TestList_showsVersionAndWorld(t *testing.T) {
	t.Parallel()

	client := &stubClient{profiles: ProfilesData{
		Active: profAlice,
		Profiles: []profile.Profile{
			{Name: profAlice, Label: profAliceLabel, Version: "14.0.0", World: "avalon"},
		},
	}}
	resp := &stubResponder{}
	if err := makeCommands(client).List(context.Background(), resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"14.0.0", "avalon"} {
		if !strings.Contains(resp.content, want) {
			t.Errorf("expected %q in list, got %q", want, resp.content)
		}
	}
}

func TestList_clientError(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		resp := &stubResponder{}
		client := &stubClient{status: StatusData{Online: true, Active: profBob}}
		if err := makeCommands(
			client,
		).Switch(context.Background(), resp, profBob, false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ephemeral {
			t.Error("initial switch acknowledgement should not be ephemeral")
		}
		if !strings.Contains(resp.edited, profBob) {
			t.Errorf("expected profile name in edited response, got %q", resp.edited)
		}
		if !strings.Contains(resp.edited, "✅") {
			t.Errorf("expected success marker in edited response, got %q", resp.edited)
		}
	})
}

func TestSwitch_serverNeverReturns_warns(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		resp := &stubResponder{}
		client := &stubClient{statusErr: errors.New("connection refused")}
		if err := makeCommands(
			client,
		).Switch(context.Background(), resp, profBob, false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(resp.edited, "⚠️") {
			t.Errorf("expected a warning when the server never answers, got %q", resp.edited)
		}
	})
}

func TestSwitch_failure_editsMessage(t *testing.T) {
	t.Parallel()

	client := &stubClient{switchErr: errors.New("unknown profile")}
	resp := &stubResponder{}
	if err := makeCommands(client).Switch(context.Background(), resp, "nobody", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.edited, "❌") {
		t.Errorf("expected failure marker in edited response, got %q", resp.edited)
	}
	if !strings.Contains(resp.edited, "unknown profile") {
		t.Errorf("expected error detail in edited response, got %q", resp.edited)
	}
}

func TestSwitch_passesForce(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		client := &stubClient{status: StatusData{Online: true, Active: profBob}}
		resp := &stubResponder{}
		if err := makeCommands(
			client,
		).Switch(context.Background(), resp, profBob, true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !client.gotForce {
			t.Error("expected force flag to be forwarded to the client")
		}
	})
}

func TestVersions_listsInstalled(t *testing.T) {
	t.Parallel()

	client := &stubClient{
		versions: VersionsData{Active: "14.361.0", Installed: []string{"14.361.0", "13.351.0"}},
	}
	resp := &stubResponder{}
	if err := makeCommands(client).Versions(context.Background(), resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.content, "14.361.0") || !strings.Contains(resp.content, "▶") {
		t.Errorf("expected active version marked, got %q", resp.content)
	}
}

func TestVersions_empty(t *testing.T) {
	t.Parallel()

	resp := &stubResponder{}
	if err := makeCommands(&stubClient{}).Versions(context.Background(), resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.content, "No Foundry versions") {
		t.Errorf("expected empty message, got %q", resp.content)
	}
}

func TestDownload_success(t *testing.T) {
	t.Parallel()

	client := &stubClient{}
	resp := &stubResponder{}
	if err := makeCommands(
		client,
	).Download(context.Background(), resp, "14.361.0", "https://signed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.gotVersion != "14.361.0" || client.gotURL != "https://signed" {
		t.Errorf("download args not forwarded: %q %q", client.gotVersion, client.gotURL)
	}
	if !strings.Contains(resp.edited, "✅") {
		t.Errorf("expected success marker, got %q", resp.edited)
	}
}

func TestDownload_failureRelaysError(t *testing.T) {
	t.Parallel()

	client := &stubClient{downloadErr: errors.New("no source for 9.9.9")}
	resp := &stubResponder{}
	if err := makeCommands(client).Download(context.Background(), resp, "9.9.9", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.edited, "❌") || !strings.Contains(resp.edited, "no source") {
		t.Errorf("expected relayed error, got %q", resp.edited)
	}
}

func TestLogs_showsLines(t *testing.T) {
	t.Parallel()

	client := &stubClient{logs: LogsData{Lines: []string{"line one", "line two"}}}
	resp := &stubResponder{}
	if err := makeCommands(client).Logs(context.Background(), resp, 20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.gotTail != 20 {
		t.Errorf("expected tail 20, got %d", client.gotTail)
	}
	if !strings.Contains(resp.content, "line one") || !strings.Contains(resp.content, "```") {
		t.Errorf("expected code-block with lines, got %q", resp.content)
	}
}

func TestLogs_empty(t *testing.T) {
	t.Parallel()

	resp := &stubResponder{}
	if err := makeCommands(&stubClient{}).Logs(context.Background(), resp, 20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.content, "No logs") {
		t.Errorf("expected empty message, got %q", resp.content)
	}
}

func TestStatus_offline(t *testing.T) {
	t.Parallel()

	client := &stubClient{status: StatusData{Active: profAlice, Version: "14.0.0", Online: false}}
	resp := &stubResponder{}
	if err := makeCommands(client).Status(context.Background(), resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.content, profAlice) || !strings.Contains(resp.content, "14.0.0") {
		t.Errorf("expected active+version in response, got %q", resp.content)
	}
	if !strings.Contains(resp.content, "offline") {
		t.Errorf("expected offline marker, got %q", resp.content)
	}
}

func TestStatus_online(t *testing.T) {
	t.Parallel()

	client := &stubClient{status: StatusData{
		Active:        profAlice,
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

func TestSwitchDoesNotConfirmAgainstTheOutgoingProcess(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// The manager marks the new profile active before spawning the child, so a
		// server that never restarted would otherwise look like a successful switch.
		stillRunning := StatusData{Online: true, Active: profBob, UptimeMS: 90_000}
		client := stubClient{status: stillRunning}
		resp := &stubResponder{}

		if err := makeCommands(&client).Switch(
			context.Background(), resp, profBob, false,
		); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(resp.edited, "⚠️") {
			t.Fatalf("edited = %q, want a warning: the process never cycled", resp.edited)
		}
	})
}

func TestSwitchConfirmsOnceTheProcessCycled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		client := stubClient{statusSeq: []StatusData{
			{Online: true, Active: profAlice, UptimeMS: 90_000},
			{Online: false},
			{Online: true, Active: profBob, UptimeMS: 1_000},
		}}
		resp := &stubResponder{}

		if err := makeCommands(&client).Switch(
			context.Background(), resp, profBob, false,
		); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(resp.edited, "✅") {
			t.Fatalf("edited = %q, want it confirmed", resp.edited)
		}
	})
}
