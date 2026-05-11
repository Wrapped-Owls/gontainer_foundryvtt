package forge

import (
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

// Event is the marker interface for observer notifications. Concrete
// event types carry their payload as exported fields.
type Event interface {
	EventKind() string
}

// EventResolved fires after Resolve returns a Plan.
type EventResolved struct {
	Plan Plan
}

func (EventResolved) EventKind() string { return "resolved" }

// EventInstalling fires before a Source.Materialise call.
type EventInstalling struct {
	Source source.Source
	Target string
}

func (EventInstalling) EventKind() string { return "installing" }

// EventInstalled fires after Acquire returns a final install.
type EventInstalled struct {
	Install Install
}

func (EventInstalled) EventKind() string { return "installed" }

// EventSkipped fires when Acquire short-circuits (reusing a candidate).
type EventSkipped struct {
	Reason string
	Install Install
}

func (EventSkipped) EventKind() string { return "skipped" }

// Observer receives every fourcery event. Implementations must not
// panic on unknown event types; use a type switch.
type Observer interface {
	Notify(Event)
}

// SlogObserver is the default observer; logs each event at Info via
// the wrapped slog.Logger.
type SlogObserver struct {
	Logger *slog.Logger
}

// Notify implements Observer.
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

// noopObserver is used when no observer is configured.
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
