package forge

import (
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

type Action int

const (
	ActionUseExisting Action = iota + 1
	ActionInstallFromSource
)

type Plan struct {
	Action          Action
	Candidate       *Candidate
	Source          source.Source
	TargetRoot      string
	ResolvedVersion version.Version
}

type Install struct {
	Root    string
	Version version.Version
}
