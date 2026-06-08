package executor

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunStreamsOutputAndSucceeds(t *testing.T) {
	var events []Event
	result := Run(context.Background(), Spec{
		Command: "printf 'hello\\n'",
	}, func(event Event) {
		events = append(events, event)
	})

	if result.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %s: %v", result.Status, result.Err)
	}
	if result.ExitCode == nil || *result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %#v", result.ExitCode)
	}
	if len(events) != 1 || events[0].Message != "hello" {
		t.Fatalf("expected streamed hello event, got %#v", events)
	}
}

func TestRunReportsFailedExitCode(t *testing.T) {
	result := Run(context.Background(), Spec{
		Command: "exit 7",
	}, func(Event) {})

	if result.Status != "failed" {
		t.Fatalf("expected failed, got %s", result.Status)
	}
	if result.ExitCode == nil || *result.ExitCode != 7 {
		t.Fatalf("expected exit code 7, got %#v", result.ExitCode)
	}
}

func TestRunTimesOut(t *testing.T) {
	result := Run(context.Background(), Spec{
		Command: "sleep 2",
		Timeout: 50 * time.Millisecond,
	}, func(Event) {})

	if result.Status != "timed_out" {
		t.Fatalf("expected timed_out, got %s: %v", result.Status, result.Err)
	}
	if result.Err == nil || !strings.Contains(result.Err.Error(), "deadline") {
		t.Fatalf("expected deadline error, got %v", result.Err)
	}
}
