package runner

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/runlet/runlet/apps/runner/internal/api"
	"github.com/runlet/runlet/apps/runner/internal/config"
	"github.com/runlet/runlet/apps/runner/internal/executor"
)

type API interface {
	RegisterRunner(context.Context, api.RegisterRunnerRequest) (api.RegisterRunnerResponse, error)
	Heartbeat(context.Context, string, api.HeartbeatRequest) error
	Claim(context.Context, string, api.ClaimRequest) (*api.RunSpec, error)
	SendRunEvent(context.Context, string, api.RunEventRequest) error
	FinishRun(context.Context, string, api.FinishRunRequest) error
}

type Options struct {
	Once bool
}

type Service struct {
	cfg       config.Config
	api       API
	logger    *slog.Logger
	currentID atomic.Value
}

func New(cfg config.Config, apiClient API, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{cfg: cfg, api: apiClient, logger: logger}
}

func (s *Service) Start(ctx context.Context, options Options) error {
	runnerID, err := s.runnerID(ctx)
	if err != nil {
		return err
	}

	s.logger.Info("runner started", "runner_id", runnerID, "name", s.cfg.Name)

	heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
	defer stopHeartbeat()
	go s.heartbeatLoop(heartbeatCtx, runnerID)

	if options.Once {
		return s.claimAndRunOnce(ctx, runnerID)
	}

	ticker := time.NewTicker(s.cfg.PollInterval())
	defer ticker.Stop()

	for {
		if err := s.claimAndRunOnce(ctx, runnerID); err != nil {
			s.logger.Warn("claim or run failed", "error", err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (s *Service) runnerID(ctx context.Context) (string, error) {
	if s.cfg.RunnerID != "" {
		s.logger.Info("using configured runner id", "runner_id", s.cfg.RunnerID)
		return s.cfg.RunnerID, nil
	}

	s.logger.Info("registering runner", "name", s.cfg.Name)
	resp, err := s.api.RegisterRunner(ctx, api.RegisterRunnerRequest{
		Name:   s.cfg.Name,
		Labels: s.cfg.Labels,
	})
	if err != nil {
		return "", err
	}
	s.logger.Info("runner registered", "runner_id", resp.RunnerID, "name", s.cfg.Name)
	return resp.RunnerID, nil
}

func (s *Service) heartbeatLoop(ctx context.Context, runnerID string) {
	ticker := time.NewTicker(s.cfg.HeartbeatInterval())
	defer ticker.Stop()

	for {
		s.sendHeartbeat(ctx, runnerID)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) sendHeartbeat(ctx context.Context, runnerID string) {
	status := "idle"
	currentRun, _ := s.currentID.Load().(string)
	if currentRun != "" {
		status = "running"
	}

	err := s.api.Heartbeat(ctx, runnerID, api.HeartbeatRequest{
		Status:     status,
		CurrentRun: currentRun,
		Labels:     s.cfg.Labels,
	})
	if err != nil {
		s.logger.Warn("heartbeat failed", "error", err)
	}
}

func (s *Service) claimAndRunOnce(ctx context.Context, runnerID string) error {
	run, err := s.api.Claim(ctx, runnerID, api.ClaimRequest{
		Capacity: 1,
		Labels:   s.cfg.Labels,
	})
	if err != nil || run == nil {
		return err
	}

	s.currentID.Store(run.ID)
	defer s.currentID.Store("")

	return s.execute(ctx, run)
}

func (s *Service) execute(ctx context.Context, run *api.RunSpec) error {
	var seq atomic.Int64
	nextSeq := func() int64 {
		return seq.Add(1)
	}

	sendEvent := func(level, stream, message string, metadata map[string]any) {
		err := s.api.SendRunEvent(ctx, run.ID, api.RunEventRequest{
			Sequence:  nextSeq(),
			Level:     level,
			Stream:    stream,
			Message:   message,
			Metadata:  metadata,
			CreatedAt: time.Now().UTC(),
		})
		if err != nil {
			s.logger.Warn("failed to send run event", "run_id", run.ID, "error", err)
		}
	}

	if run.Mode != "" && run.Mode != "shell" {
		message := fmt.Sprintf("unsupported run mode %q", run.Mode)
		sendEvent("error", "runner", message, nil)
		return s.api.FinishRun(ctx, run.ID, api.FinishRunRequest{
			Status:     "failed",
			FinishedAt: time.Now().UTC(),
			Message:    message,
		})
	}

	cwd := run.Cwd
	if cwd == "" {
		cwd = s.cfg.DefaultWorkspace
	} else if !filepath.IsAbs(cwd) && s.cfg.DefaultWorkspace != "" {
		cwd = filepath.Join(s.cfg.DefaultWorkspace, cwd)
	}
	if cwd != "" {
		if abs, err := filepath.Abs(cwd); err == nil {
			cwd = abs
		}
	}

	timeout := s.cfg.DefaultTimeout()
	if run.TimeoutSeconds > 0 {
		timeout = time.Duration(run.TimeoutSeconds) * time.Second
	}

	s.logger.Info("executing run", "run_id", run.ID, "cwd", cwd)
	sendEvent("info", "runner", "run started", map[string]any{
		"cwd":            cwd,
		"timeoutSeconds": int(timeout.Seconds()),
	})

	result := executor.Run(ctx, executor.Spec{
		Command: run.Command,
		Cwd:     cwd,
		Env:     run.Env,
		Shell:   s.cfg.Shell,
		Timeout: timeout,
	}, func(event executor.Event) {
		sendEvent("info", event.Stream, event.Message, nil)
	})

	message := "run finished"
	if result.Err != nil {
		message = result.Err.Error()
		sendEvent("error", "runner", message, nil)
	}

	err := s.api.FinishRun(ctx, run.ID, api.FinishRunRequest{
		Status:     result.Status,
		ExitCode:   result.ExitCode,
		FinishedAt: time.Now().UTC(),
		Message:    message,
	})
	if err != nil {
		return err
	}

	s.logger.Info("run finished", "run_id", run.ID, "status", result.Status)
	return nil
}
