package command

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestAwaitReady(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		client    stubClient
		wantWorld string
		wantErr   bool
	}{
		{
			name:      "an already-live server is accepted on the first probe",
			client:    stubClient{status: StatusData{Online: true, World: worldSharn}},
			wantWorld: worldSharn,
		},
		{
			name: "probing continues while the server is still down",
			client: stubClient{statusSeq: []StatusData{
				{Online: false},
				{Online: false},
				{Online: true, World: worldSharn},
			}},
			wantWorld: worldSharn,
		},
		{
			name:    "an unreachable manager is waited out, then reported",
			client:  stubClient{statusErr: errors.New("connection refused")},
			wantErr: true,
		},
		{
			name:    "a server that stays down is reported",
			client:  stubClient{status: StatusData{Online: false}},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			synctest.Test(t, func(t *testing.T) {
				client := testCase.client
				data, err := makeCommands(&client).awaitReady(
					context.Background(),
					func(s StatusData) bool { return s.Online },
				)

				if testCase.wantErr {
					if err == nil {
						t.Fatal("expected the wait to time out")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if data.World != testCase.wantWorld {
					t.Fatalf("world = %q, want %q", data.World, testCase.wantWorld)
				}
			})
		})
	}
}

func TestAwaitReadyGivesUpAfterTheConfirmWindow(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		client := stubClient{status: StatusData{Online: false}}
		started := time.Now()

		if _, err := makeCommands(&client).awaitReady(
			context.Background(),
			func(s StatusData) bool { return s.Online },
		); err == nil {
			t.Fatal("expected the wait to time out")
		}

		if waited := time.Since(started); waited < confirmWindow {
			t.Fatalf("gave up after %s, want at least %s", waited, confirmWindow)
		}
	})
}

func TestWorldSuffix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		data StatusData
		want string
	}{
		{
			name: "a booted world is named",
			data: StatusData{WorldActive: true, World: worldEberron},
			want: worldEberron,
		},
		{
			name: "no world means the setup screen",
			data: StatusData{WorldActive: false, World: worldEberron},
			want: "setup",
		},
		{
			name: "an unnamed world means the setup screen",
			data: StatusData{WorldActive: true},
			want: "setup",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := worldSuffix(testCase.data); !strings.Contains(got, testCase.want) {
				t.Fatalf("worldSuffix() = %q, want it to mention %q", got, testCase.want)
			}
		})
	}
}
