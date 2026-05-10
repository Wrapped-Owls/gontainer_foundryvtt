package backoff

import (
	"encoding/json"
	"io/fs"
	"os"
)

const filePerm fs.FileMode = 0o644

func readState(path string) (State, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, err
	}
	if s.ConsecutiveFailures < 0 {
		s.ConsecutiveFailures = 0
	}
	return s, nil
}

func writeStateAtomic(path string, s State) error {
	tmp := path + ".tmp"
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.WriteFile(tmp, b, filePerm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
