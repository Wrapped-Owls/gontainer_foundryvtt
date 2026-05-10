package config

import (
	"encoding/json"
	"io"
)

func WriteConfig(w io.Writer, c Config) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(c)
}
