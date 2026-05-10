package config

import (
	"encoding/json"
	"io"
)

// WriteConfig marshals c as indented JSON to w in the flat shape Foundry's
// options.json expects (two-space indent, no HTML escaping).
func WriteConfig(w io.Writer, c Config) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(c)
}
