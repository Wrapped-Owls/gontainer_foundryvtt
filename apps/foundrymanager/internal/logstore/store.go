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
		Kind:    "crash",
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
	start := max(cursor-s.dropped, 0)
	if start > len(s.events) {
		start = len(s.events)
	}
	out := make([]Event, len(s.events)-start)
	copy(out, s.events[start:])
	return out, next
}

// appendLine stores a line and flags it as an event when it matches a pattern.
// The caller holds s.mu.
func (s *Store) appendLine(line string) {
	s.lines = append(s.lines, line)
	if len(s.lines) > s.maxLines {
		s.lines = s.lines[len(s.lines)-s.maxLines:]
	}
	if s.matches(line) {
		s.pushEvent(Event{Time: time.Now(), Kind: "error", Message: line})
	}
}

// matches reports whether line contains any configured pattern. Caller holds s.mu.
func (s *Store) matches(line string) bool {
	if len(s.patterns) == 0 {
		return false
	}
	low := strings.ToLower(line)
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
