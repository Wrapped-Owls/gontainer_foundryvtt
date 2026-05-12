package forge

import (
	"context"
	"errors"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

// fakeSource is a test-only source whose Probe and Kind are configurable.
type fakeSource struct {
	kind    source.Kind
	version string
	probeEr error
}

func (f *fakeSource) Kind() source.Kind { return f.kind }
func (f *fakeSource) Describe() string  { return "fake-" + string(f.kind) }
func (f *fakeSource) Probe(context.Context) (string, error) {
	if f.probeEr != nil {
		return "", f.probeEr
	}
	return f.version, nil
}

func (f *fakeSource) Materialise(context.Context, string) (source.Result, error) {
	return source.Result{Kind: f.kind, Version: f.version}, nil
}

func candFor(version string) Candidate { return newCandidate("/foundry/"+version, version) }

func TestResolve_DesiredMatchesCandidate(t *testing.T) {
	r := NewResolver("/foundry")
	plan, err := r.Resolve(
		context.Background(),
		"14.361.2",
		[]Candidate{candFor("14.361.2")},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Action != ActionUseExisting {
		t.Errorf("want UseExisting, got %v", plan.Action)
	}
}

func TestResolve_DesiredMatchesSourceProbe(t *testing.T) {
	r := NewResolver("/foundry")
	srcs := []source.Source{
		&fakeSource{kind: source.KindZip, version: "14.361.2"},
	}
	plan, err := r.Resolve(context.Background(), "14.361.2", nil, srcs)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Action != ActionInstallFromSource {
		t.Fatalf("want InstallFromSource, got %v", plan.Action)
	}
	if plan.Source.Kind() != source.KindZip {
		t.Errorf("want zip source, got %v", plan.Source.Kind())
	}
	if plan.TargetRoot == "/foundry" || plan.TargetRoot == "" {
		t.Errorf("want versioned target, got %q", plan.TargetRoot)
	}
}

func TestResolve_DesiredNoLocalFallsBackToURL(t *testing.T) {
	r := NewResolver("/foundry")
	srcs := []source.Source{
		&fakeSource{kind: source.KindURL, probeEr: source.ErrVersionUnknown},
	}
	plan, err := r.Resolve(context.Background(), "14.361.2", nil, srcs)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Source.Kind() != source.KindURL {
		t.Errorf("want url, got %v", plan.Source.Kind())
	}
}

func TestResolve_DesiredNoMatchReturnsErrNoMatch(t *testing.T) {
	r := NewResolver("/foundry")
	_, err := r.Resolve(context.Background(), "14.999.0", nil, nil)
	if !errors.Is(err, source.ErrNoMatch) {
		t.Fatalf("want ErrNoMatch, got %v", err)
	}
}

func TestResolve_UndesiredPrefersURL(t *testing.T) {
	r := NewResolver("/foundry")
	srcs := []source.Source{
		&fakeSource{kind: source.KindZip, version: "14.360.0"},
		&fakeSource{kind: source.KindURL, probeEr: source.ErrVersionUnknown},
	}
	plan, err := r.Resolve(context.Background(), "", nil, srcs)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Source.Kind() != source.KindURL {
		t.Errorf("want URL (operator intent), got %v", plan.Source.Kind())
	}
}

func TestResolve_UndesiredLocalPicksHighest(t *testing.T) {
	r := NewResolver("/foundry")
	srcs := []source.Source{
		&fakeSource{kind: source.KindZip, version: "14.360.0"},
		&fakeSource{kind: source.KindFolder, version: "14.361.2"},
	}
	plan, err := r.Resolve(context.Background(), "", nil, srcs)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := plan.Source.Probe(context.Background())
	if v != "14.361.2" {
		t.Errorf("want 14.361.2 local source, got %q", v)
	}
}

func TestResolve_UndesiredFallsBackToLatestInstalled(t *testing.T) {
	r := NewResolver("/foundry")
	cands := []Candidate{candFor("14.361.0"), candFor("14.360.0")}
	plan, err := r.Resolve(context.Background(), "", cands, nil)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Action != ActionUseExisting {
		t.Fatalf("want UseExisting, got %v", plan.Action)
	}
	if plan.Candidate.Version != "14.361.0" {
		t.Errorf("want 14.361.0, got %q", plan.Candidate.Version)
	}
}

func TestResolve_DesiredMatchesURLLabel(t *testing.T) {
	r := NewResolver("/foundry")
	srcs := []source.Source{
		source.NewURL("https://example.invalid/x.zip", nil, "14.361.2", ""),
	}
	plan, err := r.Resolve(context.Background(), "14.361.2", nil, srcs)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Source == nil || plan.Source.Kind() != source.KindURL {
		t.Fatalf("want labelled URL source, got %+v", plan.Source)
	}
}

func TestResolve_LatestPicksHighestLocalSource(t *testing.T) {
	r := NewResolver("/foundry")
	srcs := []source.Source{
		&fakeSource{kind: source.KindZip, version: "14.360.0"},
		&fakeSource{kind: source.KindFolder, version: "14.361.2"},
	}
	plan, err := r.Resolve(context.Background(), VersionLatest, nil, srcs)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := plan.Source.Probe(context.Background())
	if v != "14.361.2" {
		t.Errorf("want 14.361.2, got %q", v)
	}
}

func TestResolve_LatestPrefersLocalOverURL(t *testing.T) {
	r := NewResolver("/foundry")
	srcs := []source.Source{
		&fakeSource{kind: source.KindZip, version: "14.361.2"},
		&fakeSource{kind: source.KindURL, probeEr: source.ErrVersionUnknown},
	}
	plan, err := r.Resolve(context.Background(), VersionLatest, nil, srcs)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Source.Kind() != source.KindZip {
		t.Errorf("want zip (local-first), got %v", plan.Source.Kind())
	}
}

func TestResolve_LatestFallsBackToCandidate(t *testing.T) {
	r := NewResolver("/foundry")
	cands := []Candidate{candFor("14.361.0"), candFor("14.360.0")}
	plan, err := r.Resolve(context.Background(), VersionLatest, cands, nil)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Action != ActionUseExisting {
		t.Fatalf("want UseExisting, got %v", plan.Action)
	}
	if plan.Candidate.Version != "14.361.0" {
		t.Errorf("want 14.361.0, got %q", plan.Candidate.Version)
	}
}

func TestResolve_LatestFallsBackToURL(t *testing.T) {
	r := NewResolver("/foundry")
	srcs := []source.Source{
		&fakeSource{kind: source.KindURL, probeEr: source.ErrVersionUnknown},
	}
	plan, err := r.Resolve(context.Background(), VersionLatest, nil, srcs)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Source.Kind() != source.KindURL {
		t.Errorf("want URL fallback, got %v", plan.Source.Kind())
	}
}

func TestResolve_LatestNoLocalNoRemote(t *testing.T) {
	r := NewResolver("/foundry")
	_, err := r.Resolve(context.Background(), VersionLatest, nil, nil)
	if !errors.Is(err, source.ErrNoMatch) {
		t.Fatalf("want ErrNoMatch, got %v", err)
	}
}

func TestVersionsEqual(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"14.361.2", "14.361.2", true},
		{"14.361.2", "14.361", true},
		{"14.361.0", "14.361.2", false}, // desired carries patch, requires exact
		{"14.361.2", "14.362.0", false},
		{"v14.361", "14.361", true},
		{"", "", true},
		{"14.361", "", false},
	}
	for _, c := range cases {
		got := versionsEqual(c.a, c.b)
		if got != c.want {
			t.Errorf("versionsEqual(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}
