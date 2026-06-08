package api

import "time"

type RegisterRunnerRequest struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type RegisterRunnerResponse struct {
	RunnerID string `json:"runnerId"`
}

type HeartbeatRequest struct {
	Status     string            `json:"status"`
	CurrentRun string            `json:"currentRunId,omitempty"`
	Labels     map[string]string `json:"labels"`
}

type ClaimRequest struct {
	Capacity int               `json:"capacity"`
	Labels   map[string]string `json:"labels"`
}

type ClaimResponse struct {
	Run *RunSpec `json:"run"`
}

type RunSpec struct {
	ID             string            `json:"id"`
	RunletID       string            `json:"runletId"`
	Mode           string            `json:"mode"`
	Command        string            `json:"command"`
	Cwd            string            `json:"cwd"`
	Env            map[string]string `json:"env"`
	TimeoutSeconds int               `json:"timeoutSeconds"`
}

type RunEventRequest struct {
	Sequence  int64          `json:"sequence"`
	Level     string         `json:"level"`
	Stream    string         `json:"stream,omitempty"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

type FinishRunRequest struct {
	Status     string    `json:"status"`
	ExitCode   *int      `json:"exitCode"`
	FinishedAt time.Time `json:"finishedAt"`
	Message    string    `json:"message,omitempty"`
}
