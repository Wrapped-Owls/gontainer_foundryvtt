package logstore

import (
	"strings"
	"testing"
)

func TestWrite_retainsOnlyErrorLines(t *testing.T) {
	s := New(10, 10, nil)
	if _, err := s.Write([]byte("info: started\nError: boom\ndebug: tick\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := s.Tail(0)
	if len(got) != 1 || got[0] != "Error: boom" {
		t.Errorf("expected only the error line retained, got %v", got)
	}
}

func TestWrite_classifiesLine(t *testing.T) {
	testCases := []struct {
		name    string
		line    string
		isError bool
	}{
		{"plain error", "Error: something failed", true},
		{"uppercase marker", "UNCAUGHT exception", true},
		{"unhandled", "unhandled promise rejection", true},
		{"fatal", "FATAL: cannot bind", true},
		{"permission", "User Bob lacks permission to update Actor", true},
		{"info dropped", "info: world loaded", false},
		{"debug dropped", "debug: heartbeat", false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			s := New(10, 10, nil)
			if _, err := s.Write([]byte(testCase.line + "\n")); err != nil {
				t.Fatalf("write: %v", err)
			}
			events, _ := s.EventsSince(0)
			if got := len(events) == 1; got != testCase.isError {
				t.Errorf("isError(%q) = %v, want %v", testCase.line, got, testCase.isError)
			}
		})
	}
}

func TestWrite_configuredPatternRetained(t *testing.T) {
	s := New(10, 10, []string{"disk full"})
	if _, err := s.Write([]byte("warn: DISK FULL on volume\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	events, _ := s.EventsSince(0)
	if len(events) != 1 || !strings.Contains(events[0].Message, "DISK FULL") {
		t.Errorf("expected the configured pattern to be captured, got %+v", events)
	}
}

func TestWrite_coalescesDuplicateErrorEvents(t *testing.T) {
	s := New(10, 10, nil)
	line := "Error: actor lacks permission\n"
	for range 3 {
		if _, err := s.Write([]byte(line)); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if got := s.Tail(0); len(got) != 3 {
		t.Errorf("expected all 3 lines retained in the log, got %d", len(got))
	}
	events, next := s.EventsSince(0)
	if len(events) != 1 || next != 1 {
		t.Errorf("expected the flood coalesced to one event, got %d next=%d", len(events), next)
	}
}

func TestWrite_distinctErrorsNotCoalesced(t *testing.T) {
	s := New(10, 10, nil)
	if _, err := s.Write([]byte("Error: one\nError: two\nError: one\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if events, _ := s.EventsSince(0); len(events) != 3 {
		t.Errorf("expected 3 events for distinct consecutive messages, got %d", len(events))
	}
}

func TestWrite_buffersPartialErrorLine(t *testing.T) {
	s := New(10, 10, nil)
	if _, err := s.Write([]byte("Err")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write([]byte("or: late\n")); err != nil {
		t.Fatal(err)
	}
	got := s.Tail(0)
	if len(got) != 1 || got[0] != "Error: late" {
		t.Errorf("expected joined partial error line, got %v", got)
	}
}

func TestTail_evictsOldestErrorLine(t *testing.T) {
	s := New(2, 10, nil)
	if _, err := s.Write([]byte("Error: a\nError: b\nError: c\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := s.Tail(0)
	if strings.Join(got, ",") != "Error: b,Error: c" {
		t.Errorf("expected ring to keep last 2 error lines, got %v", got)
	}
}

func TestRecordCrash_andCursor(t *testing.T) {
	s := New(10, 10, nil)
	s.RecordCrash(1)
	s.RecordCrash(139)

	first, next := s.EventsSince(0)
	if len(first) != 2 || next != 2 {
		t.Fatalf("expected 2 events, got %d next=%d", len(first), next)
	}
	if first[0].Kind != "crash" {
		t.Errorf("expected crash kind, got %q", first[0].Kind)
	}
	rest, next2 := s.EventsSince(next)
	if len(rest) != 0 || next2 != 2 {
		t.Errorf("expected no new events, got %d next=%d", len(rest), next2)
	}
}

func TestEventsSince_evictionAdvancesCursor(t *testing.T) {
	s := New(10, 2, nil)
	s.RecordCrash(1)
	s.RecordCrash(2)
	s.RecordCrash(3) // evicts the first

	events, next := s.EventsSince(0)
	if len(events) != 2 || next != 3 {
		t.Fatalf("expected 2 retained events with next=3, got %d next=%d", len(events), next)
	}
	if !strings.Contains(events[0].Message, "code 2") {
		t.Errorf("expected oldest retained to be code 2, got %q", events[0].Message)
	}
}
