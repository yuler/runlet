package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientRegisterRunner(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v1/runners" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var req RegisterRunnerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Name != "local" || req.Labels["os"] != "darwin" {
			t.Fatalf("unexpected request %#v", req)
		}
		_ = json.NewEncoder(w).Encode(RegisterRunnerResponse{RunnerID: "rnr_123"})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "token")
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.RegisterRunner(context.Background(), RegisterRunnerRequest{
		Name:   "local",
		Labels: map[string]string{"os": "darwin"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RunnerID != "rnr_123" {
		t.Fatalf("unexpected runner id %s", resp.RunnerID)
	}
	if authHeader != "Bearer token" {
		t.Fatalf("unexpected authorization header %q", authHeader)
	}
}

func TestClientHeartbeat(t *testing.T) {
	var method string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		if r.URL.Path != "/api/v1/runners/rnr_123" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "token")
	if err != nil {
		t.Fatal(err)
	}

	err = client.Heartbeat(context.Background(), "rnr_123", HeartbeatRequest{Status: "idle"})
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPatch {
		t.Fatalf("unexpected method %s", method)
	}
}

func TestClientPreservesBaseURLPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/acme/api/v1/runners" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(RegisterRunnerResponse{RunnerID: "rnr_123"})
	}))
	defer server.Close()

	client, err := NewClient(server.URL+"/acme", "token")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.RegisterRunner(context.Background(), RegisterRunnerRequest{Name: "local"}); err != nil {
		t.Fatal(err)
	}
}

func TestClientClaim(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/runners/rnr_123/claims" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(ClaimResponse{Run: &RunSpec{
			ID:             "run_123",
			Mode:           "shell",
			Command:        "pwd",
			Cwd:            "subdir",
			Env:            map[string]string{"RUNLET": "1"},
			TimeoutSeconds: 30,
		}})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "token")
	if err != nil {
		t.Fatal(err)
	}

	run, err := client.Claim(context.Background(), "rnr_123", ClaimRequest{Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	if run == nil {
		t.Fatal("expected run")
	}
	if run.Mode != "shell" || run.Command != "pwd" || run.Cwd != "subdir" {
		t.Fatalf("unexpected run %#v", run)
	}
}

func TestClientReturnsHTTPErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "token")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.RegisterRunner(context.Background(), RegisterRunnerRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClientSendRunEvent(t *testing.T) {
	var path string
	var got RunEventRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "token")
	if err != nil {
		t.Fatal(err)
	}

	err = client.SendRunEvent(context.Background(), "run_123", RunEventRequest{
		Sequence: 1,
		Level:    "info",
		Stream:   "stdout",
		Message:  "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/runs/run_123/events" {
		t.Fatalf("unexpected path %s", path)
	}
	if got.Message != "hello" || got.Sequence != 1 {
		t.Fatalf("unexpected event payload: %#v", got)
	}
}

func TestClientFinishRun(t *testing.T) {
	var method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "token")
	if err != nil {
		t.Fatal(err)
	}

	exitCode := 0
	err = client.FinishRun(context.Background(), "run_123", FinishRunRequest{
		Status:   "succeeded",
		ExitCode: &exitCode,
	})
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPost {
		t.Fatalf("unexpected method %s", method)
	}
	if path != "/api/v1/runs/run_123/finish" {
		t.Fatalf("unexpected path %s", path)
	}
}

func TestNewClientRejectsInvalidURL(t *testing.T) {
	if _, err := NewClient("not-a-url", "token"); err == nil {
		t.Fatal("expected error for missing scheme/host")
	}
}

func TestNewClientStripsTrailingSlash(t *testing.T) {
	client, err := NewClient("http://example.com/acme/", "token")
	if err != nil {
		t.Fatal(err)
	}
	if client.baseURL != "http://example.com/acme" {
		t.Fatalf("expected trimmed base url, got %q", client.baseURL)
	}
}
