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

func TestStartOnceClaimsAndExecutesQueuedRun(t *testing.T) {
	workspace := t.TempDir()
	fake := &fakeAPI{
		runs: []*api.RunSpec{{
			ID:      "run_xyz",
			Mode:    "shell",
			Command: "printf hi",
		}},
	}
	service := New(config.Config{
		APIURL:                   "http://localhost:3000",
		Token:                    "t",
		Name:                     "local",
		Concurrency:              1,
		DefaultWorkspace:         workspace,
		PollIntervalSeconds:      1,
		HeartbeatIntervalSeconds: 60,
		DefaultTimeoutSeconds:    30,
	}, fake, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := service.Start(ctx, Options{Once: true}); err != nil {
		t.Fatal(err)
	}

	if fake.register != 1 {
		t.Fatalf("expected runner to register once, got %d", fake.register)
	}
	if fake.finish.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %#v", fake.finish)
	}

	var stdoutEvent api.RunEventRequest
	for _, ev := range fake.events {
		if ev.Stream == "stdout" {
			stdoutEvent = ev
			break
		}
	}
	if stdoutEvent.Message != "hi" {
		t.Fatalf("expected stdout 'hi' event, got events %#v", fake.events)
	}
}

func TestStartOnceWithConfiguredRunnerSkipsRegistration(t *testing.T) {
	fake := &fakeAPI{}
	service := New(config.Config{
		APIURL:                   "http://localhost:3000",
		Token:                    "t",
		RunnerID:                 "rnr_existing",
		Name:                     "local",
		Concurrency:              1,
		PollIntervalSeconds:      1,
		HeartbeatIntervalSeconds: 60,
		DefaultTimeoutSeconds:    30,
	}, fake, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := service.Start(ctx, Options{Once: true}); err != nil {
		t.Fatal(err)
	}
	if fake.register != 0 {
		t.Fatalf("expected no registration when runner id is set, got %d", fake.register)
	}
}

type fakeAPI struct {
	events   []api.RunEventRequest
	finish   api.FinishRunRequest
	runs     []*api.RunSpec
	register int
}

func (f *fakeAPI) RegisterRunner(context.Context, api.RegisterRunnerRequest) (api.RegisterRunnerResponse, error) {
	f.register++
	return api.RegisterRunnerResponse{RunnerID: "rnr_fake"}, nil
}

func (f *fakeAPI) Heartbeat(context.Context, string, api.HeartbeatRequest) error {
	return nil
}

func (f *fakeAPI) Claim(context.Context, string, api.ClaimRequest) (*api.RunSpec, error) {
	if len(f.runs) == 0 {
		return nil, nil
	}
	run := f.runs[0]
	f.runs = f.runs[1:]
	return run, nil
}

func (f *fakeAPI) SendRunEvent(_ context.Context, _ string, req api.RunEventRequest) error {
	f.events = append(f.events, req)
	return nil
}

func (f *fakeAPI) FinishRun(_ context.Context, _ string, req api.FinishRunRequest) error {
	f.finish = req
	return nil
}
