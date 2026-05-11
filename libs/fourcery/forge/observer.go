package forge

import (
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

type Event interface {
	EventKind() string
}

type EventResolved struct {
	Plan Plan
}

func (EventResolved) EventKind() string { return "resolved" }

type EventInstalling struct {
	Source source.Source
	Target string
}

func (EventInstalling) EventKind() string { return "installing" }

type EventInstalled struct {
	Install Install
}

func (EventInstalled) EventKind() string { return "installed" }

type EventSkipped struct {
	Reason  string
	Install Install
}

func (EventSkipped) EventKind() string { return "skipped" }

type Observer interface {
	Notify(Event)
}

type SlogObserver struct {
	Logger *slog.Logger
}

func (o SlogObserver) Notify(e Event) {
	if o.Logger == nil {
		return
	}
	switch ev := e.(type) {
	case EventResolved:
		o.Logger.Info(
			"forge resolved",
			"action", actionString(ev.Plan.Action),
			"version", ev.Plan.ResolvedVersion,
			"target", ev.Plan.TargetRoot,
		)
	case EventInstalling:
		o.Logger.Info(
			"forge installing",
			"source", ev.Source.Describe(),
			"kind", string(ev.Source.Kind()),
			"target", ev.Target,
		)
	case EventInstalled:
		o.Logger.Info(
			"forge installed",
			"root", ev.Install.Root,
			"version", ev.Install.Version,
		)
	case EventSkipped:
		o.Logger.Info(
			"forge skipped",
			"reason", ev.Reason,
			"root", ev.Install.Root,
			"version", ev.Install.Version,
		)
	}
}

type noopObserver struct{}

func (noopObserver) Notify(Event) {}

func actionString(a Action) string {
	switch a {
	case ActionUseExisting:
		return "use-existing"
	case ActionInstallFromSource:
		return "install-from-source"
	default:
		return "unknown"
	}
}
