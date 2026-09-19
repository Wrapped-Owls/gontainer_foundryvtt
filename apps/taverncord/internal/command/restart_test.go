package command

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"
)

func TestRestart(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		client       stubClient
		interrupt    Interrupt
		wantMarker   string
		wantRestarts int
	}{
		{
			name: "a cycled process is confirmed back online",
			client: stubClient{statusSeq: []StatusData{
				{Online: true, UptimeMS: 90_000},
				{Online: false},
				{Online: true, UptimeMS: 1_000, WorldActive: true, World: worldEberron},
			}},
			wantMarker:   "✅",
			wantRestarts: 1,
		},
		{
			name: "an unchanged uptime is not accepted as a restart",
			client: stubClient{statusSeq: []StatusData{
				{Online: true, UptimeMS: 90_000},
			}},
			wantMarker:   "⚠",
			wantRestarts: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			synctest.Test(t, func(t *testing.T) {
				client := testCase.client
				resp := &stubResponder{}

				if err := makeCommands(&client).Restart(
					context.Background(), resp, testCase.interrupt,
				); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(resp.edited, testCase.wantMarker) {
					t.Fatalf("edited = %q, want it to contain %q", resp.edited, testCase.wantMarker)
				}
				if client.restarts != testCase.wantRestarts {
					t.Fatalf("restart calls = %d, want %d", client.restarts, testCase.wantRestarts)
				}
				if client.gotInterrupt != testCase.interrupt {
					t.Fatalf(
						"interrupt forwarded as %q, want %q",
						client.gotInterrupt,
						testCase.interrupt,
					)
				}
			})
		})
	}
}

func TestRestartAServerThatWasAlreadyDown(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		client := stubClient{statusSeq: []StatusData{
			{Online: false},
			{Online: true, UptimeMS: 2_000},
		}}
		resp := &stubResponder{}

		if err := makeCommands(
			&client,
		).Restart(context.Background(), resp, InterruptWhenIdle); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(resp.edited, "✅") {
			t.Fatalf("edited = %q, want it confirmed back online", resp.edited)
		}
	})
}

func TestRestartFailures(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		client     stubClient
		interrupt  Interrupt
		wantMarker string
	}{
		{
			name:       "a refused restart reports the reason",
			client:     stubClient{restartErr: errors.New("2 user(s) currently online")},
			wantMarker: "❌",
		},
		{
			name:       "interrupt is forwarded to the manager",
			client:     stubClient{status: StatusData{Online: true}},
			interrupt:  InterruptAlways,
			wantMarker: "✅",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			synctest.Test(t, func(t *testing.T) {
				client := testCase.client
				resp := &stubResponder{}

				if err := makeCommands(&client).Restart(
					context.Background(), resp, testCase.interrupt,
				); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(resp.edited, testCase.wantMarker) {
					t.Fatalf("edited = %q, want %q", resp.edited, testCase.wantMarker)
				}
				if client.gotInterrupt != testCase.interrupt {
					t.Fatalf(
						"interrupt forwarded as %q, want %q",
						client.gotInterrupt,
						testCase.interrupt,
					)
				}
			})
		})
	}
}

func TestRestartAcknowledgementIsPublic(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		client := stubClient{status: StatusData{Online: true}}
		resp := &stubResponder{}

		if err := makeCommands(
			&client,
		).Restart(context.Background(), resp, InterruptWhenIdle); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.visibility != Public {
			t.Error("a restart affects everyone connected, so the notice must not be ephemeral")
		}
	})
}
