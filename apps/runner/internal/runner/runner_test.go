package runner

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/runlet/runlet/apps/runner/internal/api"
	"github.com/runlet/runlet/apps/runner/internal/config"
)

func TestExecuteUsesWorkspaceRelativeCwd(t *testing.T) {
	workspace := t.TempDir()
	if err := os.Mkdir(filepath.Join(workspace, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	fake := &fakeAPI{}
	service := New(config.Config{
		DefaultWorkspace:      workspace,
		DefaultTimeoutSeconds: 30,
	}, fake, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	err := service.execute(context.Background(), &api.RunSpec{
		ID:      "run_123",
		Mode:    "shell",
		Command: "pwd",
		Cwd:     "subdir",
	})

	if err != nil {
		t.Fatal(err)
	}
	if fake.finish.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %#v", fake.finish)
	}
	want := filepath.Join(workspace, "subdir")
	want, err = filepath.EvalSymlinks(want)
	if err != nil {
		t.Fatal(err)
	}
	if got := fake.events[1].Message; got != want {
		t.Fatalf("expected pwd output %q, got %q", want, got)
	}
}

func TestExecuteRejectsUnsupportedMode(t *testing.T) {
	fake := &fakeAPI{}
	service := New(config.Config{DefaultTimeoutSeconds: 30}, fake, nil)

	err := service.execute(context.Background(), &api.RunSpec{
		ID:   "run_123",
		Mode: "docker",
	})

	if err != nil {
		t.Fatal(err)
	}
	if fake.finish.Status != "failed" {
		t.Fatalf("expected failed finish, got %#v", fake.finish)
	}
	if fake.events[0].Level != "error" {
		t.Fatalf("expected error event, got %#v", fake.events)
	}
}

type fakeAPI struct {
	events []api.RunEventRequest
	finish api.FinishRunRequest
}

func (f *fakeAPI) RegisterRunner(context.Context, api.RegisterRunnerRequest) (api.RegisterRunnerResponse, error) {
	return api.RegisterRunnerResponse{}, nil
}

func (f *fakeAPI) Heartbeat(context.Context, string, api.HeartbeatRequest) error {
	return nil
}

func (f *fakeAPI) Claim(context.Context, string, api.ClaimRequest) (*api.RunSpec, error) {
	return nil, nil
}

func (f *fakeAPI) SendRunEvent(_ context.Context, _ string, req api.RunEventRequest) error {
	f.events = append(f.events, req)
	return nil
}

func (f *fakeAPI) FinishRun(_ context.Context, _ string, req api.FinishRunRequest) error {
	f.finish = req
	return nil
}
