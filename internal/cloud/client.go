// Package cloud is a client for the Deca Agent Jobs API.
//
// It mirrors the HTTP contract used by the opencode-action-harness
// runner (see artifacts/harness/runner.py): worker calls authenticate
// with the X-Worker-Token header, user calls authenticate with a
// Bearer token. The base URL and credentials always come from the
// environment (DECA_AGENT_API_BASE_URL, DECA_API_KEY,
// DECA_AGENT_WORKER_TOKEN) so secrets never land in files or logs.
package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to the Agent Jobs API rooted at baseURL, e.g.
// https://<genlabsapi-host> with paths under /deca-agents/v1/jobs.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userToken  string
	workerTok  string
}

// NewClient builds a client. Either token may be empty if that auth
// path is not needed.
func NewClient(baseURL, userToken, workerToken string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userToken:  userToken,
		workerTok:  workerToken,
	}
}

func (c *Client) do(ctx context.Context, method, path string, worker bool, payload any) ([]byte, int, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to encode request payload: %w", err)
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to build request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "star-cloud-probe")
	if worker {
		if c.workerTok == "" {
			return nil, 0, fmt.Errorf("Worker token is required but not set (DECA_AGENT_WORKER_TOKEN)")
		}
		req.Header.Set("X-Worker-Token", c.workerTok)
	} else {
		if c.userToken == "" {
			return nil, 0, fmt.Errorf("API key is required but not set (DECA_API_KEY)")
		}
		req.Header.Set("Authorization", "Bearer "+c.userToken)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("Request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to read response: %w", err)
	}
	return respBody, resp.StatusCode, nil
}

// Ping checks reachability and auth against the jobs collection. The
// exact user-side list endpoint lives in genlabsapi; this tries the
// documented worker-adjacent paths and reports what the server says,
// which is the whole point of the proof: Star -> API round trip.
func (c *Client) Ping(ctx context.Context) (string, error) {
	for _, path := range []string{"/deca-agents/v1/jobs", "/deca/v1/models"} {
		body, status, err := c.do(ctx, http.MethodGet, path, false, nil)
		if err != nil {
			return "", err
		}
		if status < 500 {
			return fmt.Sprintf("GET %s -> HTTP %d: %s", path, status, truncate(string(body), 500)), nil
		}
	}
	return "", fmt.Errorf("API base did not answer on known paths")
}

// JobStatus fetches one job as the worker would (X-Worker-Token).
func (c *Client) JobStatus(ctx context.Context, jobID string) (string, error) {
	body, status, err := c.do(ctx, http.MethodGet, "/deca-agents/v1/jobs/"+jobID, true, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("HTTP %d: %s", status, truncate(string(body), 1000)), nil
}

// PatchStatus mirrors runner.py patch_status.
func (c *Client) PatchStatus(ctx context.Context, jobID, status string) (int, error) {
	_, code, err := c.do(ctx, http.MethodPatch, "/deca-agents/v1/jobs/"+jobID, true, map[string]any{"status": status})
	return code, err
}

// PostEvents mirrors runner.py flush_events.
func (c *Client) PostEvents(ctx context.Context, jobID string, events []any) (int, error) {
	_, code, err := c.do(ctx, http.MethodPost, "/deca-agents/v1/jobs/"+jobID+"/events", true, map[string]any{"events": events})
	return code, err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
