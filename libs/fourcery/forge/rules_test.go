package forge

import (
	"context"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

// runRules

func TestRunRules_FirstSucceeds(t *testing.T) {
	t.Parallel()

	want := Plan{Action: ActionUseExisting, ResolvedVersion: "14.361.2"}
	first := func(_ context.Context, _ []Candidate, _ []source.Source) (Plan, bool) {
		return want, true
	}
	second := func(_ context.Context, _ []Candidate, _ []source.Source) (Plan, bool) {
		t.Error("second rule must not be called when first succeeds")
		return Plan{}, false
	}

	got, ok := runRules(context.Background(), nil, nil, []rule{first, second})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestRunRules_SecondSucceeds(t *testing.T) {
	t.Parallel()

	want := Plan{Action: ActionUseExisting, ResolvedVersion: "14.361.2"}
	skip := func(_ context.Context, _ []Candidate, _ []source.Source) (Plan, bool) { return Plan{}, false }
	hit := func(_ context.Context, _ []Candidate, _ []source.Source) (Plan, bool) { return want, true }

	got, ok := runRules(context.Background(), nil, nil, []rule{skip, hit})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestRunRules_AllFail(t *testing.T) {
	t.Parallel()

	skip := func(_ context.Context, _ []Candidate, _ []source.Source) (Plan, bool) { return Plan{}, false }
	_, ok := runRules(context.Background(), nil, nil, []rule{skip, skip})
	if ok {
		t.Fatal("expected ok=false when all rules fail")
	}
}

// ruleUseMatchingCandidate

func TestRuleUseMatchingCandidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		desired    string
		candidates []Candidate
		wantOk     bool
		wantVer    string
	}{
		{
			name:       "exact match",
			desired:    "14.361.2",
			candidates: []Candidate{candFor("14.361.2")},
			wantOk:     true,
			wantVer:    "14.361.2",
		},
		{
			name:       "no candidates",
			desired:    "14.361.2",
			candidates: nil,
			wantOk:     false,
		},
		{
			name:       "version mismatch",
			desired:    "14.999.0",
			candidates: []Candidate{candFor("14.361.2")},
			wantOk:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := ruleUseMatchingCandidate(tt.desired)
			plan, ok := r(context.Background(), tt.candidates, nil)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if plan.Action != ActionUseExisting {
				t.Errorf("action = %v, want UseExisting", plan.Action)
			}
			if plan.ResolvedVersion != tt.wantVer {
				t.Errorf("version = %q, want %q", plan.ResolvedVersion, tt.wantVer)
			}
		})
	}
}

// ruleMatchingSource

func TestRuleMatchingSource(t *testing.T) {
	t.Parallel()

	r := NewResolver("/foundry")
	tests := []struct {
		name    string
		desired string
		sources []source.Source
		wantOk  bool
		wantKnd source.Kind
	}{
		{
			name:    "probe matches",
			desired: "14.361.2",
			sources: []source.Source{&fakeSource{kind: source.KindZip, version: "14.361.2"}},
			wantOk:  true,
			wantKnd: source.KindZip,
		},
		{
			name:    "probe mismatch",
			desired: "14.999.0",
			sources: []source.Source{&fakeSource{kind: source.KindZip, version: "14.361.2"}},
			wantOk:  false,
		},
		{
			name:    "probe errors are skipped",
			desired: "14.361.2",
			sources: []source.Source{&fakeSource{kind: source.KindZip, probeEr: source.ErrVersionUnknown}},
			wantOk:  false,
		},
		{
			name:    "no sources",
			desired: "14.361.2",
			sources: nil,
			wantOk:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plan, ok := ruleMatchingSource(r, tt.desired)(context.Background(), nil, tt.sources)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && plan.Source.Kind() != tt.wantKnd {
				t.Errorf("kind = %v, want %v", plan.Source.Kind(), tt.wantKnd)
			}
		})
	}
}

// ruleUnknownVersionSource

func TestRuleUnknownVersionSource(t *testing.T) {
	t.Parallel()

	r := NewResolver("/foundry")
	tests := []struct {
		name    string
		sources []source.Source
		wantOk  bool
	}{
		{
			name:    "url with unknown version",
			sources: []source.Source{&fakeSource{kind: source.KindURL, probeEr: source.ErrVersionUnknown}},
			wantOk:  true,
		},
		{
			name:    "source with known version is skipped",
			sources: []source.Source{&fakeSource{kind: source.KindZip, version: "14.361.2"}},
			wantOk:  false,
		},
		{
			name:    "no sources",
			sources: nil,
			wantOk:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plan, ok := ruleUnknownVersionSource(r, "14.361.2")(context.Background(), nil, tt.sources)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && plan.Action != ActionInstallFromSource {
				t.Errorf("action = %v, want InstallFromSource", plan.Action)
			}
		})
	}
}

// ruleHighestLocalSource

func TestRuleHighestLocalSource(t *testing.T) {
	t.Parallel()

	r := NewResolver("/foundry")
	tests := []struct {
		name    string
		sources []source.Source
		wantOk  bool
		wantVer string
	}{
		{
			name: "picks highest among zip and folder",
			sources: []source.Source{
				&fakeSource{kind: source.KindZip, version: "14.360.0"},
				&fakeSource{kind: source.KindFolder, version: "14.361.2"},
			},
			wantOk:  true,
			wantVer: "14.361.2",
		},
		{
			name: "only remote sources returns false",
			sources: []source.Source{
				&fakeSource{kind: source.KindURL, probeEr: source.ErrVersionUnknown},
			},
			wantOk: false,
		},
		{
			name: "local with probe error is skipped",
			sources: []source.Source{
				&fakeSource{kind: source.KindZip, probeEr: source.ErrVersionUnknown},
				&fakeSource{kind: source.KindFolder, version: "14.360.0"},
			},
			wantOk:  true,
			wantVer: "14.360.0",
		},
		{
			name:    "no sources",
			sources: nil,
			wantOk:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plan, ok := ruleHighestLocalSource(r)(context.Background(), nil, tt.sources)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if plan.ResolvedVersion != tt.wantVer {
				t.Errorf("version = %q, want %q", plan.ResolvedVersion, tt.wantVer)
			}
		})
	}
}

// ruleLatestCandidate

func TestRuleLatestCandidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		candidates []Candidate
		wantOk     bool
		wantVer    string
	}{
		{
			name:       "picks first (newest) candidate",
			candidates: []Candidate{candFor("14.361.2"), candFor("14.360.0")},
			wantOk:     true,
			wantVer:    "14.361.2",
		},
		{
			name:       "no candidates",
			candidates: nil,
			wantOk:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plan, ok := ruleLatestCandidate()(context.Background(), tt.candidates, nil)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if plan.Action != ActionUseExisting {
				t.Errorf("action = %v, want UseExisting", plan.Action)
			}
			if plan.ResolvedVersion != tt.wantVer {
				t.Errorf("version = %q, want %q", plan.ResolvedVersion, tt.wantVer)
			}
		})
	}
}

// ruleFirstSourceOfKind

func TestRuleFirstSourceOfKind(t *testing.T) {
	t.Parallel()

	r := NewResolver("/foundry")
	tests := []struct {
		name    string
		kind    source.Kind
		sources []source.Source
		wantOk  bool
	}{
		{
			name:    "matching kind found",
			kind:    source.KindURL,
			sources: []source.Source{&fakeSource{kind: source.KindURL, probeEr: source.ErrVersionUnknown}},
			wantOk:  true,
		},
		{
			name:    "no matching kind",
			kind:    source.KindSession,
			sources: []source.Source{&fakeSource{kind: source.KindURL, probeEr: source.ErrVersionUnknown}},
			wantOk:  false,
		},
		{
			name:    "no sources",
			kind:    source.KindURL,
			sources: nil,
			wantOk:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plan, ok := ruleFirstSourceOfKind(r, tt.kind)(context.Background(), nil, tt.sources)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && plan.Action != ActionInstallFromSource {
				t.Errorf("action = %v, want InstallFromSource", plan.Action)
			}
		})
	}
}
