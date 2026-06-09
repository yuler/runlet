package executor

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRunStreamsOutputAndSucceeds(t *testing.T) {
	var events []Event
	result := Run(context.Background(), Spec{
		Command: "printf 'hello\\nworld\\n'",
	}, func(event Event) {
		events = append(events, event)
	})

	if result.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %s: %v", result.Status, result.Err)
	}
	if result.ExitCode == nil || *result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %#v", result.ExitCode)
	}
	if len(events) != 2 || events[0].Message != "hello" || events[1].Message != "world" {
		t.Fatalf("expected streamed hello/world events, got %#v", events)
	}
}

func TestRunDoesNotDropOutputLines(t *testing.T) {
	const lines = 200
	var got []Event
	result := Run(context.Background(), Spec{
		Command: fmt.Sprintf("for i in $(seq 1 %d); do echo line-$i; done", lines),
	}, func(event Event) {
		got = append(got, event)
	})

	if result.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %s: %v", result.Status, result.Err)
	}
	if len(got) != lines {
		t.Fatalf("expected %d events, got %d", lines, len(got))
	}
	for i, ev := range got {
		want := fmt.Sprintf("line-%d", i+1)
		if ev.Message != want {
			t.Fatalf("event %d: expected %q, got %q", i, want, ev.Message)
		}
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

func TestRunFailsOnEmptyCommand(t *testing.T) {
	result := Run(context.Background(), Spec{}, func(Event) {})

	if result.Status != "failed" {
		t.Fatalf("expected failed, got %s", result.Status)
	}
	if result.Err == nil || !strings.Contains(result.Err.Error(), "command") {
		t.Fatalf("expected command error, got %v", result.Err)
	}
}

func TestRunInjectsExtraEnv(t *testing.T) {
	var captured []string
	result := Run(context.Background(), Spec{
		Command: "echo RUNLET_TEST=$RUNLET_TEST",
		Env:     map[string]string{"RUNLET_TEST": "hello"},
	}, func(event Event) {
		captured = append(captured, event.Message)
	})

	if result.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %s: %v", result.Status, result.Err)
	}
	if len(captured) != 1 || captured[0] != "RUNLET_TEST=hello" {
		t.Fatalf("expected RUNLET_TEST=hello, got %#v", captured)
	}
}

func TestRunIsCanceledByContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	result := Run(ctx, Spec{
		Command: "sleep 2",
	}, func(Event) {})

	if result.Status != "canceled" {
		t.Fatalf("expected canceled, got %s: %v", result.Status, result.Err)
	}
}

func TestRunRespectsCustomShell(t *testing.T) {
	var events []Event
	result := Run(context.Background(), Spec{
		Command: "echo from-bash",
		Shell:   "/bin/sh",
	}, func(event Event) {
		events = append(events, event)
	})

	if result.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %s: %v", result.Status, result.Err)
	}
	if len(events) != 1 || events[0].Message != "from-bash" {
		t.Fatalf("expected from-bash event, got %#v", events)
	}
}

func TestMergeEnvAppendsToBase(t *testing.T) {
	base := []string{"BASE=1"}
	merged := mergeEnv(base, map[string]string{"EXTRA": "2"})

	if len(merged) != 2 || merged[0] != "BASE=1" || merged[1] != "EXTRA=2" {
		t.Fatalf("unexpected merged env: %#v", merged)
	}
}

func TestMergeEnvReturnsBaseWhenNoExtras(t *testing.T) {
	base := []string{"BASE=1"}
	merged := mergeEnv(base, nil)

	if &merged[0] != &base[0] {
		t.Fatalf("expected mergeEnv to return the same base slice when extras are empty")
	}
}
