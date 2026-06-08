package main

import (
	"path/filepath"
	"testing"

	"github.com/runlet/runlet/apps/runner/internal/config"
)

func TestSetupDoesNotCarryStoredRunnerID(t *testing.T) {
	clearRunletEnv(t)
	path := filepath.Join(t.TempDir(), "runner.conf")
	oldSeed := config.DefaultSeed()
	oldSeed.RunnerID = "old-runner-id"
	oldSeed.Token = "old-token"
	if err := config.SaveSeed(path, oldSeed); err != nil {
		t.Fatal(err)
	}

	err := runSetup([]string{
		"new-token",
		"--api-url", "http://localhost:3000/acme",
		"--config", path,
	})
	if err != nil {
		t.Fatal(err)
	}

	seed, err := config.LoadSeed(path)
	if err != nil {
		t.Fatal(err)
	}
	if seed.RunnerID != "" {
		t.Fatalf("expected setup to clear stored runner id, got %q", seed.RunnerID)
	}
	if seed.Token != "new-token" {
		t.Fatalf("expected new token, got %q", seed.Token)
	}
	if seed.APIURL != "http://localhost:3000/acme" {
		t.Fatalf("expected new api url, got %q", seed.APIURL)
	}
}

func clearRunletEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"RUNLET_API_URL",
		"RUNLET_TOKEN",
		"RUNLET_RUNNER_ID",
		"RUNLET_RUNNER_NAME",
		"RUNLET_CONCURRENCY",
		"RUNLET_POLL_INTERVAL_SECONDS",
		"RUNLET_HEARTBEAT_INTERVAL_SECONDS",
		"RUNLET_DEFAULT_TIMEOUT_SECONDS",
		"RUNLET_WORKSPACE",
		"RUNLET_SHELL",
		"RUNLET_LABELS",
	} {
		t.Setenv(key, "")
	}
}
