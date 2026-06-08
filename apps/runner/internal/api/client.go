package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("apiUrl must include scheme and host: %q", baseURL)
	}

	return &Client{
		baseURL: strings.TrimRight(parsed.String(), "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Client) RegisterRunner(ctx context.Context, req RegisterRunnerRequest) (RegisterRunnerResponse, error) {
	var out RegisterRunnerResponse
	err := c.do(ctx, http.MethodPost, "/api/v1/runners", req, &out)
	return out, err
}

func (c *Client) Heartbeat(ctx context.Context, runnerID string, req HeartbeatRequest) error {
	return c.do(ctx, http.MethodPatch, "/api/v1/runners/"+url.PathEscape(runnerID), req, nil)
}

func (c *Client) Claim(ctx context.Context, runnerID string, req ClaimRequest) (*RunSpec, error) {
	var out ClaimResponse
	if err := c.do(ctx, http.MethodPost, "/api/v1/runners/"+url.PathEscape(runnerID)+"/claims", req, &out); err != nil {
		return nil, err
	}
	return out.Run, nil
}

func (c *Client) SendRunEvent(ctx context.Context, runID string, req RunEventRequest) error {
	return c.do(ctx, http.MethodPost, "/api/v1/runs/"+url.PathEscape(runID)+"/events", req, nil)
}

func (c *Client) FinishRun(ctx context.Context, runID string, req FinishRunRequest) error {
	return c.do(ctx, http.MethodPost, "/api/v1/runs/"+url.PathEscape(runID)+"/finish", req, nil)
}

func (c *Client) do(ctx context.Context, method, path string, in any, out any) error {
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		message := strings.TrimSpace(string(data))
		if message == "" {
			message = resp.Status
		}
		return fmt.Errorf("%s %s failed: %s", method, path, message)
	}

	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return nil
}
