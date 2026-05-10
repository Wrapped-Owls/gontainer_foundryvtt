package lifecycle

import (
	"fmt"
	"strings"
)

type InstallAction int

const (
	ActionNone InstallAction = iota
	ActionInstall
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

func DecideInstall(installed InstalledInfo, desiredVersion string) InstallAction {
	desired := strings.TrimSpace(desiredVersion)
	switch {
	case !installed.Present:
		return ActionInstall
	case desired == "":
		return ActionNone
	case installed.Version == "" || installed.Version == desired:
		return ActionNone
	default:
		return ActionUpgrade
	}
}
