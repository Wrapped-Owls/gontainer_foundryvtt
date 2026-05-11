package forge

import (
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

// Action enumerates the outcomes the resolver can produce.
type Action int

const (
	// ActionUseExisting reuses a candidate already installed.
	ActionUseExisting Action = iota + 1
	// ActionInstallFromSource installs from the chosen Source into
	// the plan's TargetRoot.
	ActionInstallFromSource
)

// Plan is the resolver's verdict. Fields populated depend on Action.
type Plan struct {
	Action          Action
	Candidate       *Candidate     // ActionUseExisting
	Source          source.Source  // ActionInstallFromSource
	TargetRoot      string         // ActionInstallFromSource
	ResolvedVersion string         // best-effort known version
}

// Install is fourcery's name for "the install we settled on": its
// final absolute path and the version detected inside it.
type Install struct {
	Root    string
	Version string
}
