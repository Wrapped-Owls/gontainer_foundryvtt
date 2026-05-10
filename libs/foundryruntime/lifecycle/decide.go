package lifecycle

import (
	"fmt"
	"strings"
)

// InstallAction enumerates what DecideInstall instructs the controller
// to perform for the upcoming run.
type InstallAction int

const (
	// ActionNone means the install on disk already matches the desired
	// version; nothing to do.
	ActionNone InstallAction = iota
	// ActionInstall means there is no Foundry on disk; perform a fresh
	// download + extract.
	ActionInstall
	// ActionUpgrade means a different version is installed; replace it.
	ActionUpgrade
)

func (a InstallAction) String() string {
	switch a {
	case ActionNone:
		return "none"
	case ActionInstall:
		return "install"
	case ActionUpgrade:
		return "upgrade"
	}
	return fmt.Sprintf("action(%d)", int(a))
}

// DecideInstall computes the install action.
//
// desiredVersion may be empty: in that case the function is permissive —
// any installed version is accepted, and a fresh install is needed only
// when nothing is present at all.
func DecideInstall(info InstalledInfo, desiredVersion string) InstallAction {
	desired := strings.TrimSpace(desiredVersion)
	switch {
	case !info.Present:
		return ActionInstall
	case desired == "":
		return ActionNone
	case info.Version == "" || info.Version == desired:
		// Either we can't read a version (assume it's fine) or it
		// already matches. Both → no action.
		return ActionNone
	default:
		return ActionUpgrade
	}
}
