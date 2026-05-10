package secfuse

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
)

func Load(path string) (Result, error) {
	if path == "" {
		path = DefaultSecretPath
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Result{}, nil
		}
		return Result{}, fmt.Errorf("secfuse: read %s: %w", path, err)
	}
	raw := map[string]any{}
	if err = json.Unmarshal(b, &raw); err != nil {
		return Result{}, fmt.Errorf("secfuse: parse %s: %w", path, err)
	}

	res := Result{SourcePath: path}
	for k, v := range raw {
		envName, ok := KnownKeys[k]
		if !ok {
			res.Unknown = append(res.Unknown, k)
			continue
		}
		s, ok := v.(string)
		if !ok || s == "" {
			continue
		}
		if err = os.Setenv(envName, s); err != nil {
			return res, fmt.Errorf("secfuse: setenv %s: %w", envName, err)
		}
		res.Applied = append(res.Applied, envName)
	}
	sort.Strings(res.Applied)
	sort.Strings(res.Unknown)
	return res, nil
}
