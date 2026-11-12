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
	"os"
	"strings"
	"time"

	"github.com/coder/websocket"
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

// sessionIDFromEnv reports the Deca web session id when Star holds one.
// Wiring the real login that mints it is future work; for now it comes
// from the environment like every other credential here.
func sessionIDFromEnv() string {
	return strings.TrimSpace(os.Getenv("DECA_SESSION_ID"))
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
	// User auth is either a Bearer API key (DECA_API_KEY) or a Deca web
	// session id (DECA_SESSION_ID) sent as the sessionid cookie — the
	// same two paths prod's terminal routes accept.
	if !worker {
		if sid := sessionIDFromEnv(); sid != "" {
			req.AddCookie(&http.Cookie{Name: "sessionid", Value: sid})
		}
	}
	req.Header.Set("User-Agent", "star-cloud-probe")
	if worker {
		if c.workerTok == "" {
			return nil, 0, fmt.Errorf("Worker token is required but not set (DECA_AGENT_WORKER_TOKEN)")
		}
		req.Header.Set("X-Worker-Token", c.workerTok)
	} else {
		hasCookie := sessionIDFromEnv() != ""
		if c.userToken == "" && !hasCookie {
			return nil, 0, fmt.Errorf("API key is required but not set (DECA_API_KEY or DECA_SESSION_ID)")
		}
		if c.userToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.userToken)
		}
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

// TerminalSession mirrors the prod terminal relay session create:
// POST /deca/v1/terminal/session with the user's Bearer key returns
// the session id plus the browser WS URL to watch it on.
func (c *Client) CreateTerminalSession(ctx context.Context) (sessionID, relayURL string, err error) {
	body, status, err := c.do(ctx, http.MethodPost, "/deca/v1/terminal/session", false, map[string]any{})
	if err != nil {
		return "", "", err
	}
	if status != http.StatusOK {
		return "", "", fmt.Errorf("Session create failed: HTTP %d: %s", status, truncate(string(body), 300))
	}
	var out struct {
		SessionID string `json:"session_id"`
		RelayURL  string `json:"relay_url"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", "", fmt.Errorf("Failed to decode session response: %w", err)
	}
	if out.SessionID == "" || out.RelayURL == "" {
		return "", "", fmt.Errorf("Session create returned no session: %s", truncate(string(body), 300))
	}
	return out.SessionID, out.RelayURL, nil
}

// WatchTerminal attaches to a terminal relay browser URL and copies the
// agent's PTY bytes to out, like watching the prod relay's TV screen.
// The runner frames PTY output as a leading type byte (0x00 = data,
// 0x02 = exit, 0x03 = ping, mirroring the harness protocol); the type
// byte is stripped before writing. It returns when the relay closes the
// stream or ctx ends.
func (c *Client) WatchTerminal(ctx context.Context, relayURL string, out io.Writer) error {
	conn, _, err := websocket.Dial(ctx, relayURL, nil)
	if err != nil {
		return fmt.Errorf("Failed to connect to relay: %w", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 {
				return nil
			}
			return fmt.Errorf("Relay read failed: %w", err)
		}
		if len(data) == 0 {
			continue
		}
		payload := data
		if data[0] <= 0x03 {
			if data[0] == 0x02 {
				fmt.Fprintf(out, "\n[agent exited: code %d]\n", exitCodeOf(data))
				return nil
			}
			if data[0] == 0x03 {
				continue
			}
			payload = data[1:]
		}
		if _, err := out.Write(payload); err != nil {
			return err
		}
	}
}

func exitCodeOf(frame []byte) int {
	if len(frame) > 1 {
		return int(frame[1])
	}
	return 0
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
