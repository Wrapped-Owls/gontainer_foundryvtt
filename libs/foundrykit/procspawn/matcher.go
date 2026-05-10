// Package procspawn manages child process execution with a clean environment.
// It:
//
//   - scrubs the child environment using a Matcher passlist (default:
//     HOME, NODE_*, TZ)
//   - executes a child process with that clean environment
//   - forwards SIGTERM / SIGINT to the child's process group
//   - waits for the child and propagates its exit status
//
// The package is intentionally tiny and depends only on the standard
// library so that the runtime binary can pull it in at near-zero cost.
package procspawn

import "strings"

// Matcher reports whether an environment variable key passes the filter.
type Matcher interface {
	Match(key string) bool
}

// ExactMatch matches a single env key by exact case-sensitive equality.
func ExactMatch(key string) Matcher { return exactMatcher{key} }

// PrefixMatch matches any key that starts with prefix.
func PrefixMatch(prefix string) Matcher { return prefixMatcher{prefix} }

// SuffixMatch matches any key that ends with suffix.
func SuffixMatch(suffix string) Matcher { return suffixMatcher{suffix} }

type exactMatcher struct{ key string }

func (m exactMatcher) Match(k string) bool { return k == m.key }

type prefixMatcher struct{ prefix string }

func (m prefixMatcher) Match(k string) bool { return strings.HasPrefix(k, m.prefix) }

type suffixMatcher struct{ suffix string }

func (m suffixMatcher) Match(k string) bool { return strings.HasSuffix(k, m.suffix) }

// DefaultPasslist is the set of environment variables forwarded to the child process.
var DefaultPasslist = []Matcher{
	ExactMatch("HOME"),
	PrefixMatch("NODE_"),
	ExactMatch("TZ"),
}
