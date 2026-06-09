package daemon

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestRunningReturnsFalseForMissingPIDFile(t *testing.T) {
	pid, running := Running(filepath.Join(t.TempDir(), "missing.pid"))
	if running {
		t.Fatalf("expected not running, got pid %d", pid)
	}
}

func TestStartWritesPIDAndLog(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("daemon start is only exercised on unix in this suite")
	}

	dir := t.TempDir()
	pidPath := filepath.Join(dir, "runner.pid")
	logPath := filepath.Join(dir, "runner.log")

	pid, err := Start(Options{
		Executable: "/bin/sh",
		Args:       []string{"-c", "sleep 2"},
		PIDPath:    pidPath,
		LogPath:    logPath,
	})
	if err != nil {
		t.Fatal(err)
	}

	storedPID, running := Running(pidPath)
	if !running {
		t.Fatal("expected background process to be running")
	}
	if storedPID != pid {
		t.Fatalf("expected stored pid %d, got %d", pid, storedPID)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		t.Fatal(err)
	}
	_ = process.Kill()
	_, _ = process.Wait()
	time.Sleep(100 * time.Millisecond)

	if _, running := Running(pidPath); running {
		t.Fatal("expected process to stop after kill")
	}
}

func TestStartRejectsDuplicateRunner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("daemon start is only exercised on unix in this suite")
	}

	dir := t.TempDir()
	pidPath := filepath.Join(dir, "runner.pid")
	logPath := filepath.Join(dir, "runner.log")

	pid, err := Start(Options{
		Executable: "/bin/sh",
		Args:       []string{"-c", "sleep 2"},
		PIDPath:    pidPath,
		LogPath:    logPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		process, findErr := os.FindProcess(pid)
		if findErr == nil {
			_ = process.Kill()
		}
	}()

	if _, err := Start(Options{
		Executable: "/bin/sh",
		Args:       []string{"-c", "sleep 2"},
		PIDPath:    pidPath,
		LogPath:    logPath,
	}); err == nil {
		t.Fatal("expected duplicate start to fail")
	}
}
