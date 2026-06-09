package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type Options struct {
	Executable string
	Args       []string
	PIDPath    string
	LogPath    string
	WorkDir    string
}

func Start(opts Options) (int, error) {
	if pid, running := Running(opts.PIDPath); running {
		return pid, fmt.Errorf("runner already running (pid %d)", pid)
	}

	if err := os.MkdirAll(filepath.Dir(opts.PIDPath), 0o700); err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(opts.LogPath), 0o700); err != nil {
		return 0, err
	}

	logFile, err := os.OpenFile(opts.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return 0, err
	}
	defer logFile.Close()

	cmd := exec.Command(opts.Executable, opts.Args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}
	cmd.SysProcAttr = detachProcessAttr()

	if err := cmd.Start(); err != nil {
		return 0, err
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile(opts.PIDPath, []byte(strconv.Itoa(pid)+"\n"), 0o600); err != nil {
		_ = cmd.Process.Kill()
		return 0, err
	}

	return pid, nil
}

func Running(pidPath string) (int, bool) {
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0, false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return 0, false
	}
	return pid, true
}
