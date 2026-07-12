package logstore

import (
	"strings"
	"testing"
)

func TestWrite_splitsLinesAndTails(t *testing.T) {
	s := New(3, 10, nil)
	if _, err := s.Write([]byte("a\nb\nc\nd\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := s.Tail(0)
	if strings.Join(got, ",") != "b,c,d" {
		t.Errorf("expected ring to keep last 3, got %v", got)
	}
}

func TestWrite_buffersPartialLine(t *testing.T) {
	s := New(10, 10, nil)
	if _, err := s.Write([]byte("hel")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write([]byte("lo\n")); err != nil {
		t.Fatal(err)
	}
	got := s.Tail(0)
	if len(got) != 1 || got[0] != "hello" {
		t.Errorf("expected joined partial line, got %v", got)
	}
}

func TestWrite_recordsEventOnPatternMatch(t *testing.T) {
	s := New(10, 10, []string{"lacks permission"})
	line := "User Bob LACKS PERMISSION to update Actor\n"
	if _, err := s.Write([]byte(line)); err != nil {
		t.Fatal(err)
	}
	events, next := s.EventsSince(0)
	if len(events) != 1 || next != 1 {
		t.Fatalf("expected one event, got %d next=%d", len(events), next)
	}
	if events[0].Kind != "error" || !strings.Contains(events[0].Message, "LACKS PERMISSION") {
		t.Errorf("unexpected event: %+v", events[0])
	}
}

func TestWrite_noEventWithoutPattern(t *testing.T) {
	s := New(10, 10, []string{"lacks permission"})
	if _, err := s.Write([]byte("all good here\n")); err != nil {
		t.Fatal(err)
	}
	if events, _ := s.EventsSince(0); len(events) != 0 {
		t.Errorf("expected no events, got %d", len(events))
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
	// Polling from the returned cursor yields nothing new.
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
