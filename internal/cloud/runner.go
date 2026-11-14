// Package cloud is a client for the Deca Agent Jobs API.
//
// RunCloudSession drives one cloud turn end to end: it creates a cloud
// ACP session, relays ACP frames, and translates them into the same
// message-store writes the local coordinator makes (assistant text,
// tool-call parts, tool-result messages, finish markers). The TUI
// renders those writes through its normal pipeline, so a cloud turn
// looks exactly like a local one.
package cloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/crush/internal/agent/notify"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/coder/acp-go-sdk"
	"github.com/coder/websocket"
)

// promptResultID is the JSON-RPC id the runner uses for session/prompt.
const promptResultID = 3

// errRunnerNotReady means no readiness frame arrived in time; the caller
// proceeds anyway so a runner without the handshake still works.
var errRunnerNotReady = errors.New("cloud runner did not announce readiness")

// RunOptions configures one cloud turn.
type RunOptions struct {
	// BaseURL is the Agent Jobs API base, e.g. https://<genlabsapi-host>.
	BaseURL string
	// UserToken is the Bearer user key (DECA_API_KEY).
	UserToken string
	// DisplayModel and DisplayProvider label the assistant message,
	// mirroring what the local coordinator records.
	DisplayModel    string
	DisplayProvider string
	// Model, ModelBaseURL, ModelAPIKey tell the cloud runner where to
	// get inference from. They come from the environment
	// (DECA_MODEL, DECA_MODEL_BASE_URL, DECA_MODEL_API_KEY) so secrets
	// stay out of files and logs.
	Model        string
	ModelBaseURL string
	ModelAPIKey  string
}

// RunOptionsFromEnv builds RunOptions from the environment. It returns
// an error naming the missing variable when the cloud path is not
// configured.
func RunOptionsFromEnv(displayModel, displayProvider string) (RunOptions, error) {
	opts := RunOptions{
		BaseURL:         strings.TrimSpace(os.Getenv("DECA_AGENT_API_BASE_URL")),
		UserToken:       os.Getenv("DECA_API_KEY"),
		DisplayModel:    displayModel,
		DisplayProvider: displayProvider,
		Model:           strings.TrimSpace(os.Getenv("DECA_MODEL")),
		ModelBaseURL:    strings.TrimSpace(os.Getenv("DECA_MODEL_BASE_URL")),
		ModelAPIKey:     os.Getenv("DECA_MODEL_API_KEY"),
	}
	if opts.BaseURL == "" {
		return opts, fmt.Errorf("Cloud mode needs DECA_AGENT_API_BASE_URL (see `star cloud status`)")
	}
	if opts.UserToken == "" && strings.TrimSpace(os.Getenv("DECA_SESSION_ID")) == "" {
		return opts, fmt.Errorf("Cloud mode needs DECA_API_KEY or DECA_SESSION_ID")
	}
	if opts.Model == "" || opts.ModelBaseURL == "" || opts.ModelAPIKey == "" {
		return opts, fmt.Errorf("Cloud mode needs DECA_MODEL, DECA_MODEL_BASE_URL and DECA_MODEL_API_KEY for cloud inference")
	}
	return opts, nil
}

// RunCloudSession runs one cloud turn for a Star session: it stores the
// user message, streams the cloud agent's ACP frames into the message
// store, then publishes the terminal RunComplete the TUI uses to retire
// the turn. It returns when the turn ends or ctx is canceled.
func (c *Client) RunCloudSession(ctx context.Context, msgs message.Service, runCompletions *pubsub.Broker[notify.RunComplete], starSessionID, prompt string, attachments []message.Attachment, opts RunOptions) error {
	if msgs == nil {
		return fmt.Errorf("Cloud mode needs the local message store")
	}

	// Store the user message first so the pending echo is replaced by
	// the stored message, exactly like a local turn.
	userParts := []message.ContentPart{message.TextContent{Text: prompt}}
	for _, a := range attachments {
		userParts = append(userParts, message.BinaryContent{Path: a.FilePath, MIMEType: a.MimeType, Data: a.Content})
	}
	if _, err := msgs.Create(ctx, starSessionID, message.CreateMessageParams{Role: message.User, Parts: userParts}); err != nil {
		return fmt.Errorf("Failed to store prompt: %w", err)
	}

	cloudSessionID, relayURL, err := c.CreateACPSession(ctx)
	if err != nil {
		return err
	}
	slog.Info("Cloud session created", "cloudSession", cloudSessionID)

	turn := &cloudTurn{msgs: msgs, sessionID: starSessionID, opts: opts}
	turnErr := c.relayTurn(ctx, turn, relayURL, prompt)

	finalizeCtx := context.Background()
	switch {
	case ctx.Err() != nil:
		turn.finish(message.FinishReasonCanceled, "User canceled request")
	case turnErr != nil:
		turn.finish(message.FinishReasonError, turnErr.Error())
	default:
		turn.finish(message.FinishReasonEndTurn, "")
	}

	if runCompletions != nil {
		rc := notify.RunComplete{SessionID: starSessionID, Cancelled: ctx.Err() != nil}
		if turnErr != nil && ctx.Err() == nil {
			rc.Error = turnErr.Error()
		}
		runCompletions.PublishMustDeliver(finalizeCtx, pubsub.UpdatedEvent, rc)
	}
	return turnErr
}

// cloudTurn owns the assistant messages for one cloud turn. The local
// coordinator writes one assistant message per model step (text and
// tool calls, then a tool result, then the next assistant message), so
// the cloud translation does the same: a completed tool call closes the
// current assistant message and the next text or tool call opens a new
// one. That keeps transcripts, renderers, and compaction identical to a
// local run.
type cloudTurn struct {
	msgs      message.Service
	sessionID string
	opts      RunOptions
	assistant *message.Message
}

// current returns the open assistant message, creating one if needed.
func (t *cloudTurn) current() (*message.Message, error) {
	if t.assistant != nil {
		return t.assistant, nil
	}
	msg, err := t.msgs.Create(context.Background(), t.sessionID, message.CreateMessageParams{
		Role:     message.Assistant,
		Parts:    []message.ContentPart{},
		Model:    t.opts.DisplayModel,
		Provider: t.opts.DisplayProvider,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create assistant message: %w", err)
	}
	t.assistant = &msg
	return t.assistant, nil
}

// finish closes the open assistant message with a terminal reason.
func (t *cloudTurn) finish(reason message.FinishReason, detail string) {
	if t.assistant == nil {
		return
	}
	msg := t.assistant
	msg.AddFinish(reason, detail, "")
	ctx := context.Background()
	if err := t.msgs.Update(ctx, *msg); err != nil {
		slog.Error("Failed to finalize cloud assistant message", "error", err)
	}
	if err := t.msgs.Flush(ctx, msg.ID); err != nil {
		slog.Error("Failed to flush cloud assistant message", "error", err)
	}
	t.assistant = nil
}

// relayTurn performs the ACP leg over the relay: config frame,
// initialize, session/new, session/prompt, then translates every
// session/update into message-store writes until the prompt result
// arrives, the runner exits, or ctx ends.
func (c *Client) relayTurn(ctx context.Context, turn *cloudTurn, relayURL, prompt string) error {
	conn, _, err := websocket.Dial(ctx, relayURL, nil)
	if err != nil {
		return fmt.Errorf("Failed to connect to cloud relay: %w", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

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

	// One goroutine reads the socket; everything else selects on the
	// resulting channels. That keeps a single reader (concurrent reads
	// on a websocket are illegal) while still allowing timeouts.
	frames := make(chan rpcMsg)
	readErr := make(chan error, 1)
	go func() {
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				readErr <- err
				return
			}
			if len(data) == 0 {
				continue
			}
			// PTY-style framing like the harness runner uses: 0x00 is
			// data, 0x02 is the runner exiting.
			if data[0] == 0x02 {
				frames <- rpcMsg{JSONRPC: "2.0", Method: "exit"}
				return
			}
			if data[0] <= 0x03 {
				data = data[1:]
			}
			if len(data) == 0 {
				continue
			}
			var msg rpcMsg
			if err := json.Unmarshal(data, &msg); err != nil {
				slog.Warn("Ignoring undecodable cloud frame", "error", err)
				continue
			}
			frames <- msg
		}
	}()

	// next returns the next frame, or an error when the turn ends or the
	// optional timeout elapses (timeout < 0 waits forever).
	next := func(timeout time.Duration) (rpcMsg, error) {
		var timeoutCh <-chan time.Time
		if timeout >= 0 {
			t := time.NewTimer(timeout)
			defer t.Stop()
			timeoutCh = t.C
		}
		select {
		case msg := <-frames:
			return msg, nil
		case err := <-readErr:
			return rpcMsg{}, err
		case <-timeoutCh:
			return rpcMsg{}, errRunnerNotReady
		case <-ctx.Done():
			return rpcMsg{}, ctx.Err()
		}
	}

	// The relay drops frames sent before the cloud runner attaches, so
	// wait for its readiness frame before configuring it. Older runners
	// never send one; fall back to sending immediately after a short
	// grace period.
	if _, err := next(3 * time.Second); err != nil && !errors.Is(err, errRunnerNotReady) {
		if !errors.Is(err, context.DeadlineExceeded) {
			return err
		}
	}

	if err := write(map[string]any{
		"type":     "acp_init",
		"model":    turn.opts.Model,
		"base_url": turn.opts.ModelBaseURL,
		"api_key":  turn.opts.ModelAPIKey,
	}); err != nil {
		return fmt.Errorf("Failed to send cloud session config: %w", err)
	}
	if err := write(rpcMsg{JSONRPC: "2.0", ID: 1, Method: "initialize",
		Params: map[string]any{"protocolVersion": 1}}); err != nil {
		return err
	}

	// waitID skips notifications, translating session/update frames
	// into the message store as they stream past.
	waitID := func(id int) (rpcMsg, error) {
		for {
			msg, err := next(-1)
			if err != nil {
				return msg, err
			}
			if msg.Method != "" {
				if msg.Method == "exit" {
					return msg, fmt.Errorf("Cloud runner exited before responding")
				}
				if msg.Method == "session/update" {
					if uerr := turn.apply(msg); uerr != nil {
						slog.Error("Failed to apply cloud update", "error", uerr)
					}
				}
				continue
			}
			if fmt.Sprint(msg.ID) == fmt.Sprint(id) {
				return msg, nil
			}
		}
	}

	if _, err := waitID(1); err != nil {
		return fmt.Errorf("Cloud ACP initialize failed: %w", err)
	}
	if err := write(rpcMsg{JSONRPC: "2.0", ID: 2, Method: "session/new",
		Params: map[string]any{"cwd": "/tmp", "mcpServers": []any{}}}); err != nil {
		return err
	}
	newResp, err := waitID(2)
	if err != nil {
		return fmt.Errorf("Cloud ACP session/new failed: %w", err)
	}
	cloudAgentSession := "cloud-1"
	if m, ok := newResp.Result.(map[string]any); ok {
		if s, ok := m["sessionId"].(string); ok && s != "" {
			cloudAgentSession = s
		}
	}
	if err := write(rpcMsg{JSONRPC: "2.0", ID: promptResultID, Method: "session/prompt",
		Params: map[string]any{
			"sessionId": cloudAgentSession,
			"prompt":    []any{map[string]any{"type": "text", "text": prompt}},
		}}); err != nil {
		return err
	}

	for {
		msg, err := next(-1)
		if err != nil {
			// ctx cancellation surfaces here; the caller finalizes.
			return err
		}
		switch {
		case msg.Method == "exit":
			return nil
		case msg.Method == "session/update":
			if uerr := turn.apply(msg); uerr != nil {
				slog.Error("Failed to apply cloud update", "error", uerr)
			}
		case msg.Method == "" && fmt.Sprint(msg.ID) == fmt.Sprint(promptResultID):
			if msg.Error != nil {
				return fmt.Errorf("Cloud prompt failed: %v", msg.Error)
			}
			return nil
		}
	}
}

// apply decodes one session/update notification with the real ACP types
// and mirrors the local coordinator's callbacks: text chunks append to
// the open assistant message, tool frames add or finish tool-call parts,
// and a completed tool call records a tool message and starts the next
// assistant message.
func (t *cloudTurn) apply(msg rpcMsg) error {
	paramsData, err := json.Marshal(msg.Params)
	if err != nil {
		return err
	}
	var notification acp.SessionNotification
	if err := json.Unmarshal(paramsData, &notification); err != nil {
		return fmt.Errorf("Failed to decode session update: %w", err)
	}
	update := notification.Update
	ctx := context.Background()

	switch {
	case update.AgentMessageChunk != nil:
		text := contentBlockText(update.AgentMessageChunk.Content)
		if text == "" {
			return nil
		}
		assistant, err := t.current()
		if err != nil {
			return err
		}
		assistant.AppendContent(text)
		return t.msgs.Update(ctx, *assistant)

	case update.AgentThoughtChunk != nil:
		text := contentBlockText(update.AgentThoughtChunk.Content)
		if text == "" {
			return nil
		}
		assistant, err := t.current()
		if err != nil {
			return err
		}
		assistant.AppendReasoningContent(text)
		return t.msgs.Update(ctx, *assistant)

	case update.ToolCall != nil:
		tc := update.ToolCall
		assistant, err := t.current()
		if err != nil {
			return err
		}
		// Keep the call unfinished until its terminal frame: a finished
		// call is folded into a summary group immediately, and a grouped
		// item can no longer receive its tool result.
		call := message.ToolCall{
			ID:   string(tc.ToolCallId),
			Name: toolCallName(tc.Title, tc.Meta),
		}
		if raw, ok := marshalRaw(tc.RawInput); ok {
			call.Input = raw
		}
		assistant.AddToolCall(call)
		return t.msgs.Update(ctx, *assistant)

	case update.ToolCallUpdate != nil:
		return t.applyToolCallUpdate(update.ToolCallUpdate)
	}
	return nil
}

// applyToolCallUpdate folds a tool_calls terminal frame into the
// transcript: the in-flight call is marked finished on the assistant
// message, the tool message records its output, and the assistant
// message is closed so the next text starts a fresh one.
func (t *cloudTurn) applyToolCallUpdate(tu *acp.SessionToolCallUpdate) error {
	ctx := context.Background()
	callID := string(tu.ToolCallId)

	var call message.ToolCall
	found := false
	if t.assistant != nil {
		for _, part := range t.assistant.Parts {
			if tc, ok := part.(message.ToolCall); ok && tc.ID == callID {
				call = tc
				found = true
				break
			}
		}
	}
	if !found {
		call = message.ToolCall{ID: callID, Name: toolCallName(strOrEmpty(tu.Title), tu.Meta)}
	}

	if tu.RawInput != nil {
		if raw, ok := marshalRaw(tu.RawInput); ok {
			call.Input = raw
		}
	}

	if tu.Status == nil || (*tu.Status != acp.ToolCallStatusCompleted && *tu.Status != acp.ToolCallStatusFailed) {
		// Still running: make sure the call is visible and unfinished.
		assistant, err := t.current()
		if err != nil {
			return err
		}
		assistant.AddToolCall(call)
		return t.msgs.Update(ctx, *assistant)
	}

	call.Finished = true
	text, _ := toolCallContentText(tu.Content)
	if text == "" {
		text = string(*tu.Status)
	}
	isErr := *tu.Status == acp.ToolCallStatusFailed

	if call.Name == "" {
		call.Name = toolCallName(strOrEmpty(tu.Title), tu.Meta)
	}
	if _, err := t.msgs.Create(ctx, t.sessionID, message.CreateMessageParams{
		Role:  message.Tool,
		Parts: []message.ContentPart{message.ToolResult{ToolCallID: callID, Name: call.Name, Content: text, IsError: isErr}},
	}); err != nil {
		return err
	}

	assistant, err := t.current()
	if err != nil {
		return err
	}
	assistant.AddToolCall(call)
	if err := t.msgs.Update(ctx, *assistant); err != nil {
		return err
	}
	// Close this step's assistant message so the next text or tool call
	// opens a new one, exactly like a local multi-step run.
	t.finish(message.FinishReasonToolUse, "")
	return nil
}

// toolCallName recovers the Star tool name for a tool frame. The bridge
// puts it in _meta (the ACP extension point) and repeats it as the
// title fallback; unknown tools still render, just without a dedicated
// renderer.
func toolCallName(title string, meta any) string {
	if m, ok := meta.(map[string]any); ok {
		if name, ok := m["starToolName"].(string); ok && name != "" {
			return name
		}
	}
	if title != "" {
		return title
	}
	return "bash"
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func marshalRaw(v any) (string, bool) {
	if v == nil {
		return "", false
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", false
	}
	return string(data), true
}

// contentBlockText extracts plain text from an ACP content block.
func contentBlockText(block acp.ContentBlock) string {
	if block.Text != nil {
		return block.Text.Text
	}
	return ""
}

// toolCallContentText extracts the text of tool result content blocks.
func toolCallContentText(content []acp.ToolCallContent) (string, bool) {
	var sb strings.Builder
	found := false
	for _, c := range content {
		if c.Content != nil {
			if t := contentBlockText(c.Content.Content); t != "" {
				sb.WriteString(t)
				found = true
			}
		}
	}
	return sb.String(), found
}
