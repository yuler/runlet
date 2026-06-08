package executor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type Spec struct {
	Command string
	Cwd     string
	Env     map[string]string
	Shell   string
	Timeout time.Duration
}

type Event struct {
	Stream  string
	Message string
}

type Result struct {
	Status   string
	ExitCode *int
	Err      error
}

func Run(ctx context.Context, spec Spec, emit func(Event)) Result {
	if spec.Command == "" {
		return Result{Status: "failed", Err: errors.New("command is empty")}
	}
	if spec.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, spec.Timeout)
		defer cancel()
	}

	cmd := shellCommand(ctx, spec.Shell, spec.Command)
	cmd.Dir = spec.Cwd
	cmd.Env = mergeEnv(os.Environ(), spec.Env)
	configureProcessGroup(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{Status: "failed", Err: err}
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return Result{Status: "failed", Err: err}
	}

	if err := cmd.Start(); err != nil {
		return Result{Status: "failed", Err: err}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go scan(&wg, stdout, "stdout", emit)
	go scan(&wg, stderr, "stderr", emit)

	waitErr := cmd.Wait()
	wg.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		terminateProcessGroup(cmd)
		return Result{Status: "timed_out", Err: ctx.Err()}
	}
	if ctx.Err() == context.Canceled {
		terminateProcessGroup(cmd)
		return Result{Status: "canceled", Err: ctx.Err()}
	}

	exitCode := 0
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			exitCode = exitErr.ExitCode()
			return Result{Status: "failed", ExitCode: &exitCode, Err: waitErr}
		}
		return Result{Status: "failed", Err: waitErr}
	}

	return Result{Status: "succeeded", ExitCode: &exitCode}
}

func shellCommand(ctx context.Context, shell, command string) *exec.Cmd {
	if shell != "" {
		return exec.CommandContext(ctx, shell, "-c", command)
	}
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd.exe", "/C", command)
	}
	return exec.CommandContext(ctx, "/bin/sh", "-c", command)
}

func mergeEnv(base []string, extra map[string]string) []string {
	if len(extra) == 0 {
		return base
	}
	env := append([]string{}, base...)
	for key, value := range extra {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

func scan(wg *sync.WaitGroup, reader io.Reader, stream string, emit func(Event)) {
	defer wg.Done()

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		emit(Event{Stream: stream, Message: scanner.Text()})
	}
	if err := scanner.Err(); err != nil {
		emit(Event{Stream: "runner", Message: fmt.Sprintf("failed to read %s: %v", stream, err)})
	}
}

func configureProcessGroup(cmd *exec.Cmd) {
	if runtime.GOOS == "windows" {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func terminateProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil || runtime.GOOS == "windows" {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}
