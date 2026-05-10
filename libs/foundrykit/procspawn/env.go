package procspawn

import "strings"

// FilterEnv returns the subset of environ ("KEY=VALUE" entries) where
// KEY matches at least one Matcher.
func FilterEnv(environ []string, matchers []Matcher) []string {
	out := make([]string, 0, len(environ))
	for _, kv := range environ {
		before, _, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		k := before
		for _, m := range matchers {
			if m.Match(k) {
				out = append(out, kv)
				break
			}
		}
	}
	return out
}
