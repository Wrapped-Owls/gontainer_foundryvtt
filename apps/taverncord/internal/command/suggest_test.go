package command

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

func profilesNamed(names ...string) ProfilesData {
	list := make([]profile.Profile, 0, len(names))
	for _, name := range names {
		list = append(list, profile.Profile{Name: name})
	}
	return ProfilesData{Profiles: list}
}

func TestSuggestProfiles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		client stubClient
		typed  string
		want   []string
	}{
		{
			name:   "nothing typed offers everything",
			client: stubClient{profiles: profilesNamed(profAlice, profBob)},
			want:   []string{profAlice, profBob},
		},
		{
			name:   "a prefix narrows the list",
			client: stubClient{profiles: profilesNamed(profAlice, profBob, "alfred")},
			typed:  "al",
			want:   []string{profAlice, "alfred"},
		},
		{
			name:   "matching ignores case",
			client: stubClient{profiles: profilesNamed(profAliceLabel, profBob)},
			typed:  "ALI",
			want:   []string{profAliceLabel},
		},
		{
			name:   "a substring matches too",
			client: stubClient{profiles: profilesNamed("master-alice", profBob)},
			typed:  profAlice,
			want:   []string{"master-alice"},
		},
		{
			name:   "no match offers nothing",
			client: stubClient{profiles: profilesNamed(profAlice)},
			typed:  "zzz",
			want:   []string{},
		},
		{
			name:   "an unreachable manager offers nothing instead of failing",
			client: stubClient{listErr: errors.New("connection refused")},
			want:   nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := testCase.client
			got := makeCommands(&client).SuggestProfiles(context.Background(), testCase.typed)
			if !slices.Equal(got, testCase.want) {
				t.Fatalf("got %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestSuggestProfilesStopsAtDiscordsCap(t *testing.T) {
	t.Parallel()

	names := make([]string, 0, maxSuggestions+10)
	for i := range maxSuggestions + 10 {
		names = append(names, "profile"+strconv.Itoa(i))
	}
	client := stubClient{profiles: profilesNamed(names...)}

	got := makeCommands(&client).SuggestProfiles(context.Background(), "")
	if len(got) != maxSuggestions {
		t.Fatalf("offered %d suggestions, want the cap of %d", len(got), maxSuggestions)
	}
}

func TestSuggestVersions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		client stubClient
		typed  string
		want   []string
	}{
		{
			name: "installed versions are offered",
			client: stubClient{
				versions: VersionsData{Installed: []string{verFoundry13, verFoundry14}},
			},
			want: []string{verFoundry13, verFoundry14},
		},
		{
			name: "a partial version narrows the list",
			client: stubClient{
				versions: VersionsData{Installed: []string{verFoundry13, verFoundry14}},
			},
			typed: "14",
			want:  []string{verFoundry14},
		},
		{
			name:   "an unreachable manager offers nothing",
			client: stubClient{versionsErr: errors.New("connection refused")},
			want:   nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := testCase.client
			got := makeCommands(&client).SuggestVersions(context.Background(), testCase.typed)
			if !slices.Equal(got, testCase.want) {
				t.Fatalf("got %v, want %v", got, testCase.want)
			}
		})
	}
}
