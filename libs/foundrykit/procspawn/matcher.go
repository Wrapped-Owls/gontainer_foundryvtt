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
	ExactMatch("LD_LIBRARY_PATH"),
}
