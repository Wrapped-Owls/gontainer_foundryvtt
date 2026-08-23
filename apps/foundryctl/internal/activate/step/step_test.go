package step

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"testing"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
)

var errStepFailed = errors.New("step: boom")

type fakeStep struct {
	label   string
	applied *[]string
	err     error
}

func (f fakeStep) Apply(_ context.Context, s *State, _ *slog.Logger) error {
	*f.applied = append(*f.applied, f.label)
	if f.err != nil {
		return f.err
	}
	s.ActiveProfile = f.label
	return nil
}

func TestRunFromShortCircuitsOnMidSequenceError(t *testing.T) {
	t.Parallel()

	var applied []string
	steps := []Step{
		fakeStep{label: "first", applied: &applied},
		fakeStep{label: "second", applied: &applied, err: errStepFailed},
		fakeStep{label: "third", applied: &applied},
	}

	got, err := RunFrom(context.Background(), discardLogger(), State{}, steps...)
	if !errors.Is(err, errStepFailed) {
		t.Fatalf("err = %v, want %v", err, errStepFailed)
	}
	if !reflect.DeepEqual(got, State{}) {
		t.Fatalf("state = %+v, want a zeroed State", got)
	}
	if len(applied) != 2 || applied[0] != "first" || applied[1] != "second" {
		t.Fatalf("applied = %v, want [first second] with the third step never reached", applied)
	}
}

func TestRunFromReturnsFinalStateOnSuccess(t *testing.T) {
	t.Parallel()

	var applied []string
	steps := []Step{
		fakeStep{label: "first", applied: &applied},
		fakeStep{label: "second", applied: &applied},
	}

	got, err := RunFrom(context.Background(), discardLogger(), State{}, steps...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ActiveProfile != "second" {
		t.Fatalf("ActiveProfile = %q, want %q", got.ActiveProfile, "second")
	}
	if len(applied) != 2 {
		t.Fatalf("applied = %v, want both steps run", applied)
	}
}

func TestRunFromStartsFromTheGivenInitialState(t *testing.T) {
	t.Parallel()

	initial := State{App: appconfig.Config{Paths: appconfig.PathsConfig{DataPath: "/data"}}}
	got, err := RunFrom(context.Background(), discardLogger(), initial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.App.Paths.DataPath != "/data" {
		t.Fatalf("DataPath = %q, want %q", got.App.Paths.DataPath, "/data")
	}
}

func TestRunStartsFromAZeroState(t *testing.T) {
	t.Parallel()

	var applied []string
	got, err := Run(context.Background(), discardLogger(), fakeStep{label: "only", applied: &applied})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ActiveProfile != "only" {
		t.Fatalf("ActiveProfile = %q, want %q", got.ActiveProfile, "only")
	}
}
