package logstore

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	DefaultBufferLines = 500
	DefaultEventBuffer = 100
)

var errorMarkers = []string{"error", "uncaught", "unhandled", "fatal", "permission"}

type Event struct {
	Time    time.Time `json:"time"`
	Kind    string    `json:"kind"`
	Message string    `json:"message"`
}

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

func New(maxLines, maxEvents int, patterns []string) *Store {
	lowered := make([]string, 0, len(patterns))
	for _, p := range patterns {
		if p = strings.ToLower(strings.TrimSpace(p)); p != "" {
			lowered = append(lowered, p)
		}
	}
	return &Store{maxLines: maxLines, maxEvents: maxEvents, patterns: lowered}
}

func (s *Store) Write(p []byte) (int, error) {
	const maxPartialLine = 64 * 1024 // bounds a newline-free stream

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

func (s *Store) RecordCrash(exitCode int) {
	const kindCrash = "crash"

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pushEvent(Event{
		Time:    time.Now(),
		Kind:    kindCrash,
		Message: fmt.Sprintf("Foundry exited with code %d", exitCode),
	})
}

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

func (s *Store) EventsSince(cursor int) ([]Event, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.dropped + len(s.events)
	start := min(max(cursor-s.dropped, 0), len(s.events))
	out := make([]Event, len(s.events)-start)
	copy(out, s.events[start:])
	return out, next
}

func (s *Store) appendLine(line string) {
	const kindError = "error"

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

func (s *Store) pushEvent(e Event) {
	s.events = append(s.events, e)
	if len(s.events) > s.maxEvents {
		drop := len(s.events) - s.maxEvents
		s.events = s.events[drop:]
		s.dropped += drop
	}
}
