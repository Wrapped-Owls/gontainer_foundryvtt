package logstore

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultBufferLines is the number of recent log lines retained.
	DefaultBufferLines = 500
	// DefaultEventBuffer is the number of detected events retained.
	DefaultEventBuffer = 100
	// maxPartialLine caps an unterminated line so a stream without newlines
	// cannot grow the buffer without bound.
	maxPartialLine = 64 * 1024
)

// errorMarkers are the built-in, case-insensitive substrings that classify a log
// line as an error worth keeping. They are matched in addition to any operator
// patterns so problem detection works out of the box; "permission" catches the
// actor-permission errors that motivated log capture.
var errorMarkers = []string{"error", "uncaught", "unhandled", "fatal", "permission"}

// Event kinds surfaced by the store.
const (
	kindError = "error"
	kindCrash = "crash"
)

// Event is a detected log occurrence worth surfacing (an error match or a crash).
type Event struct {
	Time    time.Time `json:"time"`
	Kind    string    `json:"kind"`
	Message string    `json:"message"`
}

// Store is a concurrency-safe ring buffer of log lines plus a bounded list of
// detected events. Its zero value is not usable; construct with New.
type Store struct {
	mu        sync.Mutex
	maxLines  int
	maxEvents int
	patterns  []string
	lines     []string
	partial   []byte
	events    []Event
	dropped   int
}

// New returns a Store retaining maxLines log lines and maxEvents events, flagging
// any line that contains one of patterns (matched case-insensitively).
func New(maxLines, maxEvents int, patterns []string) *Store {
	lowered := make([]string, 0, len(patterns))
	for _, p := range patterns {
		if p = strings.ToLower(strings.TrimSpace(p)); p != "" {
			lowered = append(lowered, p)
		}
	}
	return &Store{maxLines: maxLines, maxEvents: maxEvents, patterns: lowered}
}

// Write implements io.Writer: it appends complete lines to the ring buffer and
// records an event for each line matching a configured pattern. It never returns
// an error, so it is safe inside an io.MultiWriter feeding the child's stdio.
func (s *Store) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.partial = append(s.partial, p...)
	for {
		i := bytes.IndexByte(s.partial, '\n')
		if i < 0 {
			break
		}
		s.appendLine(strings.TrimRight(string(s.partial[:i]), "\r"))
		s.partial = append(s.partial[:0:0], s.partial[i+1:]...)
	}
	if len(s.partial) > maxPartialLine {
		s.appendLine(string(s.partial))
		s.partial = nil
	}
	return len(p), nil
}

// RecordCrash records a crash event for a non-zero child exit.
func (s *Store) RecordCrash(exitCode int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pushEvent(Event{
		Time:    time.Now(),
		Kind:    kindCrash,
		Message: fmt.Sprintf("Foundry exited with code %d", exitCode),
	})
}

// Tail returns the last n log lines (all of them when n <= 0 or n exceeds the
// buffer).
func (s *Store) Tail(n int) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n <= 0 || n > len(s.lines) {
		n = len(s.lines)
	}
	out := make([]string, n)
	copy(out, s.lines[len(s.lines)-n:])
	return out
}

// EventsSince returns events with a sequence at or after cursor, plus the next
// cursor to poll with. Callers start at cursor 0 and pass the returned value
// back on the following poll.
func (s *Store) EventsSince(cursor int) ([]Event, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.dropped + len(s.events)
	start := min(max(cursor-s.dropped, 0), len(s.events))
	out := make([]Event, len(s.events)-start)
	copy(out, s.events[start:])
	return out, next
}

// appendLine keeps only error lines: info/debug noise is dropped so the buffer
// and events surface problems. Each retained line becomes an error event, with
// consecutive duplicates coalesced so a flood does not spam the alert channel.
// The caller holds s.mu.
func (s *Store) appendLine(line string) {
	if !s.isError(line) {
		return
	}
	s.lines = append(s.lines, line)
	if len(s.lines) > s.maxLines {
		s.lines = s.lines[len(s.lines)-s.maxLines:]
	}
	if n := len(s.events); n > 0 {
		if last := s.events[n-1]; last.Kind == kindError && last.Message == line {
			return
		}
	}
	s.pushEvent(Event{Time: time.Now(), Kind: kindError, Message: line})
}

// isError reports whether line looks like an error: it contains a built-in error
// marker or an operator-configured pattern. Caller holds s.mu.
func (s *Store) isError(line string) bool {
	low := strings.ToLower(line)
	for _, m := range errorMarkers {
		if strings.Contains(low, m) {
			return true
		}
	}
	for _, p := range s.patterns {
		if strings.Contains(low, p) {
			return true
		}
	}
	return false
}

// pushEvent appends an event, evicting the oldest when over capacity. Caller
// holds s.mu.
func (s *Store) pushEvent(e Event) {
	s.events = append(s.events, e)
	if len(s.events) > s.maxEvents {
		drop := len(s.events) - s.maxEvents
		s.events = s.events[drop:]
		s.dropped += drop
	}
}
