package dashboard

// Switcher is the minimal interface consumed by dashboard HTTP handlers.
// Manager satisfies this interface implicitly — no import of manager/ needed.
type Switcher interface {
	RequestSwitch(name string) error
	Active() string
	Version() string
}

type profileRef struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type profilesResponse struct {
	Active   string       `json:"active"`
	Profiles []profileRef `json:"profiles"`
}

type switchBody struct {
	Profile string `json:"profile"`
}

type statusResponse struct {
	Active  string `json:"active"`
	Version string `json:"version"`
}

type errorResponse struct {
	Error string `json:"error"`
}
