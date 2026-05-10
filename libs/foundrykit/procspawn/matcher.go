package procspawn

import "strings"

type Matcher interface {
	Match(key string) bool
}

func ExactMatch(key string) Matcher { return exactMatcher{key} }

func PrefixMatch(prefix string) Matcher { return prefixMatcher{prefix} }

func SuffixMatch(suffix string) Matcher { return suffixMatcher{suffix} }

type exactMatcher struct{ key string }

func (m exactMatcher) Match(k string) bool { return k == m.key }

type prefixMatcher struct{ prefix string }

func (m prefixMatcher) Match(k string) bool { return strings.HasPrefix(k, m.prefix) }

type suffixMatcher struct{ suffix string }

func (m suffixMatcher) Match(k string) bool { return strings.HasSuffix(k, m.suffix) }

var DefaultPasslist = []Matcher{
	ExactMatch("HOME"),
	PrefixMatch("NODE_"),
	ExactMatch("TZ"),
}
