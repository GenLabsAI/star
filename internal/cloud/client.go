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
	"sync"
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

// do issues a request. apiKeyOverride, when non-empty, is used as the
// Bearer credential (e.g. a freshly enrolled model key) instead of the
// client's user token.
func (c *Client) do(ctx context.Context, method, path string, worker bool, payload any) ([]byte, int, error) {
	return c.doWith(ctx, method, path, worker, payload, "")
}

// doPublic issues a request with no user or worker credential.
func (c *Client) doPublic(ctx context.Context, method, path string, payload any) ([]byte, int, error) {
	return c.send(ctx, method, path, payload, "", false, true)
}

func (c *Client) doWith(ctx context.Context, method, path string, worker bool, payload any, apiKeyOverride string) ([]byte, int, error) {
	return c.send(ctx, method, path, payload, apiKeyOverride, worker, false)
}

func (c *Client) send(ctx context.Context, method, path string, payload any, apiKeyOverride string, worker, public bool) ([]byte, int, error) {
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
	} else if !public {
		token := c.userToken
		if apiKeyOverride != "" {
			token = apiKeyOverride
		}
		hasCookie := sessionIDFromEnv() != ""
		if token == "" && !hasCookie {
			return nil, 0, fmt.Errorf("API key is required but not set (DECA_API_KEY or DECA_SESSION_ID)")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
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
// CreateACPSession creates a cloud ACP agent session. It hits the ACP
// session route (POST /deca/v1/cloud/acp/session), which mirrors the
// terminal session flow but dispatches the ACP bridge workflow instead of
// the interactive PTY. On servers that predate the route it returns the
// 404 body as an error.
func (c *Client) CreateACPSession(ctx context.Context) (sessionID, relayURL string, err error) {
	body, status, err := c.do(ctx, http.MethodPost, "/deca/v1/cloud/acp/session", false, map[string]any{})
	if err != nil {
		return "", "", err
	}
	if status != http.StatusOK {
		return "", "", fmt.Errorf("ACP session create failed: HTTP %d: %s", status, truncate(string(body), 300))
	}
	var out struct {
		SessionID string `json:"session_id"`
		RelayURL  string `json:"relay_url"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", "", fmt.Errorf("Failed to decode ACP session response: %w", err)
	}
	if out.SessionID == "" || out.RelayURL == "" {
		return "", "", fmt.Errorf("ACP session create returned no session: %s", truncate(string(body), 300))
	}
	return out.SessionID, out.RelayURL, nil
}

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

// ACPModelConfig carries the inference access point Star sends over the relay
// so the cloud runner knows where to get inference from. The API key
// travels only in memory and over the relay frame; it is never logged.
type ACPModelConfig struct {
	Prompt  string
	Model   string
	BaseURL string
	APIKey  string
}

// rpcMsg is a minimal JSON-RPC 2.0 envelope for the ACP frames Star
// exchanges with the cloud agent.
type rpcMsg struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method,omitempty"`
	Params  any    `json:"params,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// RunACPSession performs the cloud ACP leg over a relay browser URL:
// it sends the model config frame, runs initialize/session-new/prompt,
// and writes the agent's message text to out. It returns when the
// prompt result arrives or ctx ends.
func (c *Client) RunACPSession(ctx context.Context, relayURL string, cfg ACPModelConfig, out io.Writer) error {
	conn, _, err := websocket.Dial(ctx, relayURL, nil)
	if err != nil {
		return fmt.Errorf("Failed to connect to relay: %w", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	// The relay kills a direction after ~1s with no frames (see the
	// _pump idle handling server-side), so keep this direction alive
	// with bare ping bytes while the model thinks. The runner ignores
	// non-JSON frames. Stops when the session returns.
	var wmu sync.Mutex
	done := make(chan struct{})
	defer close(done)
	go func() {
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				wmu.Lock()
				_ = conn.Write(ctx, websocket.MessageBinary, []byte{0x03})
				wmu.Unlock()
			}
		}
	}()

	write := func(v any) error {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		wmu.Lock()
		defer wmu.Unlock()
		return conn.Write(ctx, websocket.MessageText, data)
	}

	// The config frame is relay-protocol, not ACP: it tells the cloud
	// side which model endpoint and key to use for inference.
	if err := write(map[string]any{
		"type":     "acp_init",
		"prompt":   cfg.Prompt,
		"model":    cfg.Model,
		"base_url": cfg.BaseURL,
		"api_key":  cfg.APIKey,
	}); err != nil {
		return fmt.Errorf("Failed to send session config: %w", err)
	}

	if err := write(rpcMsg{JSONRPC: "2.0", ID: 1, Method: "initialize",
		Params: map[string]any{"protocolVersion": 1}}); err != nil {
		return err
	}

	read := func() (rpcMsg, error) {
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return rpcMsg{}, err
			}
			if len(data) == 0 {
				continue
			}
			// PTY-style framing like the harness runner uses: 0x00 is
			// data, 0x02 is the runner exiting.
			if data[0] == 0x02 {
				return rpcMsg{JSONRPC: "2.0", Method: "exit"}, nil
			}
			if data[0] <= 0x03 {
				data = data[1:]
			}
			if len(data) == 0 {
				continue
			}
			var msg rpcMsg
			if err := json.Unmarshal(data, &msg); err != nil {
				return rpcMsg{}, fmt.Errorf("Failed to decode ACP frame: %w", err)
			}
			return msg, nil
		}
	}

	waitID := func(id int) (rpcMsg, error) {
		for {
			msg, err := read()
			if err != nil {
				return msg, err
			}
			if msg.Method != "" {
				if text := agentChunkText(msg); text != "" {
					fmt.Fprint(out, text)
				}
				continue
			}
			if fmt.Sprint(msg.ID) == fmt.Sprint(id) {
				return msg, nil
			}
		}
	}

	if _, err := waitID(1); err != nil {
		return fmt.Errorf("ACP initialize failed: %w", err)
	}
	if err := write(rpcMsg{JSONRPC: "2.0", ID: 2, Method: "session/new",
		Params: map[string]any{"cwd": "/tmp", "mcpServers": []any{}}}); err != nil {
		return err
	}
	newResp, err := waitID(2)
	if err != nil {
		return fmt.Errorf("ACP session/new failed: %w", err)
	}
	sessionID := "cloud-1"
	if m, ok := newResp.Result.(map[string]any); ok {
		if s, ok := m["sessionId"].(string); ok && s != "" {
			sessionID = s
		}
	}
	if err := write(rpcMsg{JSONRPC: "2.0", ID: 3, Method: "session/prompt",
		Params: map[string]any{
			"sessionId": sessionID,
			"prompt":    []any{map[string]any{"type": "text", "text": cfg.Prompt}},
		}}); err != nil {
		return err
	}
	for {
		msg, err := read()
		if err != nil {
			return err
		}
		if msg.Method == "exit" {
			fmt.Fprintln(out)
			return nil
		}
		if msg.Method == "session/update" {
			if text := agentChunkText(msg); text != "" {
				fmt.Fprint(out, text)
			}
			continue
		}
		if msg.Method == "" && fmt.Sprint(msg.ID) == "3" {
			fmt.Fprintln(out)
			return nil
		}
	}
}

// agentChunkText extracts streamed agent text from a session/update frame.
func agentChunkText(msg rpcMsg) string {
	params, ok := msg.Params.(map[string]any)
	if !ok {
		return ""
	}
	update, ok := params["update"].(map[string]any)
	if !ok {
		return ""
	}
	if update["sessionUpdate"] != "agent_message_chunk" {
		return ""
	}
	content, ok := update["content"].(map[string]any)
	if !ok {
		return ""
	}
	text, _ := content["text"].(string)
	return text
}

func exitCodeOf(frame []byte) int {
	if len(frame) > 1 {
		return int(frame[1])
	}
	return 0
}

// Enroll mints a Deca opencode API key through the documented device
// enrollment endpoint. The key is returned to the caller only; it is
// never written to disk or logged.
func (c *Client) Enroll(ctx context.Context, signals map[string]string) (string, error) {
	// Enrollment is unauthenticated by design (device signal matching),
	// so it must not require a user credential.
	body, status, err := c.doPublic(ctx, http.MethodPost, "/deca/v1/opencode/enroll", map[string]any{"signals": signals})
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("Enrollment failed: HTTP %d: %s", status, truncate(string(body), 300))
	}
	var out struct {
		APIKey string `json:"api_key"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("Failed to decode enrollment response: %w", err)
	}
	if out.APIKey == "" {
		return "", fmt.Errorf("Enrollment returned no key")
	}
	return out.APIKey, nil
}

// ChatCompletion sends one prompt to the Deca model gateway and returns
// the assistant text.
func (c *Client) ChatCompletion(ctx context.Context, apiKey, model, prompt string) (string, error) {
	body, status, err := c.doWith(ctx, http.MethodPost, "/deca/v1/chat/completions", false, map[string]any{
		"model":      model,
		"messages":   []any{map[string]any{"role": "user", "content": prompt}},
		"max_tokens": 256,
		"stream":     false,
	}, apiKey)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("Completion failed: HTTP %d: %s", status, truncate(string(body), 300))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("Failed to decode completion: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("Completion returned no choices")
	}
	return out.Choices[0].Message.Content, nil
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
