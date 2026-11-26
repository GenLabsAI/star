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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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

// relayReadLimit caps one relay frame. The websocket library defaults to 32
// KiB, which is smaller than a single base64 screenshot, so an agent that
// sends an image would kill the turn before the frame could be decoded. 16 MiB
// holds a full-page screenshot at any sane resolution; anything larger is not
// a picture the transcript should swallow whole.
const relayReadLimit = 16 << 20

// initTimeout bounds the wait for the runner's initialize reply. A
// healthy runner answers instantly (no model call); silence means its
// socket is dead even if the server believed it live.
const initTimeout = 45 * time.Second

// errInitTimeout marks a runner that never answered initialize, so
// the caller can cancel it and recreate the session once.
var errInitTimeout = errors.New("Cloud runner did not answer initialize")

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
	// GitToken is an optional push credential for the cloud runner
	// (DECA_GIT_TOKEN). It travels Star -> relay -> runner over WSS
	// and is never written to files or logs.
	GitToken string
	// SetupScript and ResumeScript are the project's orb lifecycle
	// hooks (.agents/setup runs once on fresh orbs, .agents/resume
	// runs on every wake-up). They travel in the session config frame
	// and run on the ephemeral runner, never locally. SetupHash keys
	// the snapshot cache so unchanged environments restore fast.
	SetupScript  string
	SetupHash    string
	ResumeScript string
}

// RunOptionsFromEnv builds RunOptions from the environment. The model
// trio (DECA_MODEL, DECA_MODEL_BASE_URL, DECA_MODEL_API_KEY) is the
// headless fallback: callers with a selected session model (the TUI)
// fill any gaps from the selected provider's credentials first, so the
// model picker decides what the cloud runner calls. Only the relay
// base URL and user credential are required here.
func RunOptionsFromEnv(displayModel, displayProvider string) (RunOptions, error) {
	opts := RunOptions{
		BaseURL:         strings.TrimSpace(os.Getenv("DECA_AGENT_API_BASE_URL")),
		UserToken:       os.Getenv("DECA_API_KEY"),
		DisplayModel:    displayModel,
		DisplayProvider: displayProvider,
		Model:           strings.TrimSpace(os.Getenv("DECA_MODEL")),
		ModelBaseURL:    strings.TrimSpace(os.Getenv("DECA_MODEL_BASE_URL")),
		ModelAPIKey:     os.Getenv("DECA_MODEL_API_KEY"),
		GitToken:        os.Getenv("DECA_GIT_TOKEN"),
	}
	if opts.BaseURL == "" {
		return opts, fmt.Errorf("Cloud mode needs DECA_AGENT_API_BASE_URL (see `star cloud status`)")
	}
	if opts.UserToken == "" && strings.TrimSpace(os.Getenv("DECA_SESSION_ID")) == "" {
		return opts, fmt.Errorf("Cloud mode needs DECA_API_KEY or DECA_SESSION_ID")
	}
	return opts, nil
}

// NeedsModel reports whether the model trio is incomplete: the caller
// must fill it (e.g. from the selected session model) before starting
// a cloud turn.
func (o RunOptions) NeedsModel() bool {
	return o.Model == "" || o.ModelBaseURL == "" || o.ModelAPIKey == ""
}

// hookFileLimit caps how much of a lifecycle hook script travels to
// the runner inside the session config frame.
const hookFileLimit = 64 << 10

// AgentHookScripts reads the orb lifecycle hooks for a project: the
// .agents/setup script (fresh orbs) and the .agents/resume script
// (every wake-up). Missing files mean no hooks. The setup hash lets
// the runner skip straight to a matching snapshot when the
// environment has not changed.
func AgentHookScripts(projectDir string) (setup, setupHash, resume string) {
	read := func(name string) string {
		if projectDir == "" {
			return ""
		}
		data, err := os.ReadFile(filepath.Join(projectDir, ".agents", name))
		if err != nil || len(data) == 0 {
			return ""
		}
		if len(data) > hookFileLimit {
			data = data[:hookFileLimit]
		}
		return string(data)
	}
	setup = read("setup")
	if setup != "" {
		sum := sha256.Sum256([]byte(setup))
		setupHash = hex.EncodeToString(sum[:])
	}
	resume = read("resume")
	return setup, setupHash, resume
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

	// Star's own chat session id travels as the cloud chat id (the same
	// chat_session_id the jobs API uses on the classic runner path), so
	// the runner's stashed state is keyed by chat, not by turn.
	cloudSessionID, relayURL, reused, err := c.CreateACPSession(ctx, starSessionID)
	if err != nil {
		return err
	}
	slog.Info("Cloud session created", "cloudSession", cloudSessionID)

	turn := newCloudTurn(msgs, starSessionID, opts)
	turnErr := c.relayTurn(ctx, turn, relayURL, prompt)
	if errors.Is(turnErr, errInitTimeout) && reused {
		// The server believed a runner was live, but it never answered:
		// its socket is a corpse. Cancel it so it cannot attract more
		// turns, then run once more on a fresh job — which restores the
		// chat stash, so no history is lost.
		slog.Warn("Cloud reused runner did not answer; replacing it", "cloudSession", cloudSessionID)
		_ = c.DeleteSession(ctx, cloudSessionID)
		if cloudSessionID, relayURL, reused, err = c.CreateACPSession(ctx, starSessionID); err != nil {
			turnErr = err
		} else {
			slog.Info("Cloud session recreated", "cloudSession", cloudSessionID, "reused", reused)
			turnErr = c.relayTurn(ctx, turn, relayURL, prompt)
		}
	}

	// A picture the agent sent but never referenced still belongs in the
	// transcript, so link it before the turn closes.
	turn.showUnlinkedImages()

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
	// images saves the pictures the cloud agent sends and pairs them with
	// the Markdown references in its prose.
	images *imageStore
	// rawText is the unmodified text of the open assistant message and
	// textPart its index in that message, so an image arriving after the
	// prose can still turn its reference into a link.
	rawText  string
	textPart int
}

// newCloudTurn builds the translator for one turn, with an image store rooted
// at the session's cloud image directory.
func newCloudTurn(msgs message.Service, sessionID string, opts RunOptions) *cloudTurn {
	return &cloudTurn{
		msgs:      msgs,
		sessionID: sessionID,
		opts:      opts,
		images:    newImageStore(sessionID),
		textPart:  -1,
	}
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
	t.rawText = ""
	t.textPart = -1
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
	conn.SetReadLimit(relayReadLimit)

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
		"type":      "acp_init",
		"model":     turn.opts.Model,
		"base_url":  turn.opts.ModelBaseURL,
		"api_key":   turn.opts.ModelAPIKey,
		"git_token": turn.opts.GitToken,
		// Same id as the chat_session_id above: the bridge persists and
		// resumes its conversation under this key.
		"resume_key":    turn.sessionID,
		"setup_script":  turn.opts.SetupScript,
		"setup_hash":    turn.opts.SetupHash,
		"resume_script": turn.opts.ResumeScript,
	}); err != nil {
		return fmt.Errorf("Failed to send cloud session config: %w", err)
	}
	if err := write(rpcMsg{
		JSONRPC: "2.0", ID: 1, Method: "initialize",
		Params: map[string]any{"protocolVersion": 1},
	}); err != nil {
		return err
	}

	// waitID skips notifications, translating session/update frames
	// into the message store as they stream past. A non-negative
	// timeout bounds the wait: a runner that never answers is a
	// corpse, and the caller replaces it instead of hanging forever.
	waitID := func(id int, timeout time.Duration) (rpcMsg, error) {
		var timeoutCh <-chan time.Time
		if timeout >= 0 {
			t := time.NewTimer(timeout)
			defer t.Stop()
			timeoutCh = t.C
		}
		for {
			select {
			case msg := <-frames:
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
			case err := <-readErr:
				return rpcMsg{}, err
			case <-timeoutCh:
				return rpcMsg{}, errInitTimeout
			case <-ctx.Done():
				return rpcMsg{}, ctx.Err()
			}
		}
	}

	if _, err := waitID(1, initTimeout); err != nil {
		return fmt.Errorf("Cloud ACP initialize failed: %w", err)
	}
	if err := write(rpcMsg{
		JSONRPC: "2.0", ID: 2, Method: "session/new",
		Params: map[string]any{"cwd": "/tmp", "mcpServers": []any{}},
	}); err != nil {
		return err
	}
	newResp, err := waitID(2, -1)
	if err != nil {
		return fmt.Errorf("Cloud ACP session/new failed: %w", err)
	}
	cloudAgentSession := "cloud-1"
	if m, ok := newResp.Result.(map[string]any); ok {
		if s, ok := m["sessionId"].(string); ok && s != "" {
			cloudAgentSession = s
		}
	}
	if err := write(rpcMsg{
		JSONRPC: "2.0", ID: promptResultID, Method: "session/prompt",
		Params: map[string]any{
			"sessionId": cloudAgentSession,
			"prompt":    []any{map[string]any{"type": "text", "text": prompt}},
		},
	}); err != nil {
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

// appendText appends a text chunk to the open assistant message and turns any
// Markdown image reference that belongs to a picture the agent already sent
// into a clickable link. The raw text is kept so an image arriving after the
// prose can still link the reference it belongs to.
func (t *cloudTurn) appendText(delta string) error {
	assistant, err := t.current()
	if err != nil {
		return err
	}
	ctx := context.Background()
	if t.images == nil {
		assistant.AppendContent(delta)
		return t.msgs.Update(ctx, *assistant)
	}
	t.rawText += delta
	if t.textPart < 0 || t.textPart >= len(assistant.Parts) {
		assistant.Parts = append(assistant.Parts, message.TextContent{})
		t.textPart = len(assistant.Parts) - 1
	}
	text, _ := t.images.rewrite(t.rawText)
	assistant.Parts[t.textPart] = message.TextContent{Text: text}
	return t.msgs.Update(ctx, *assistant)
}

// applyImage saves one ACP image block and links the reference the agent
// already wrote for it, so the transcript shows a clickable file link even
// when the picture arrives after its prose.
func (t *cloudTurn) applyImage(img *acp.ContentBlockImage) error {
	if t.images == nil {
		return nil
	}
	if _, err := t.images.save(img); err != nil {
		return err
	}
	assistant, err := t.current()
	if err != nil {
		return err
	}
	if t.textPart < 0 || t.textPart >= len(assistant.Parts) {
		return nil
	}
	text, changed := t.images.rewrite(t.rawText)
	if !changed {
		return nil
	}
	assistant.Parts[t.textPart] = message.TextContent{Text: text}
	return t.msgs.Update(context.Background(), *assistant)
}

// saveToolImages stores image blocks a tool call returned. They stay
// unclaimed on purpose: the agent usually names the picture in the prose of a
// later step, and that reference is where the link belongs. Anything still
// unclaimed when the turn ends is shown by showUnlinkedImages.
func (t *cloudTurn) saveToolImages(content []acp.ToolCallContent) []string {
	if t.images == nil || len(content) == 0 {
		return nil
	}
	for _, c := range content {
		if c.Content == nil || c.Content.Content.Image == nil {
			continue
		}
		if _, err := t.images.save(c.Content.Content.Image); err != nil {
			slog.Warn("Failed to save cloud tool image", "error", err)
		}
	}
	return nil
}

// showUnlinkedImages appends a link for every image the turn received and the
// agent never referenced, so a picture that travelled over ACP is never
// dropped just because the prose forgot to mention it.
func (t *cloudTurn) showUnlinkedImages() {
	if t.images == nil {
		return
	}
	images := t.images.unclaimed()
	if len(images) == 0 {
		return
	}
	assistant, err := t.current()
	if err != nil {
		slog.Warn("Failed to show unlinked cloud images", "error", err)
		return
	}
	if t.textPart < 0 || t.textPart >= len(assistant.Parts) {
		assistant.Parts = append(assistant.Parts, message.TextContent{})
		t.textPart = len(assistant.Parts) - 1
	}
	current, _ := assistant.Parts[t.textPart].(message.TextContent)
	var sb strings.Builder
	sb.WriteString(current.Text)
	for _, img := range images {
		sb.WriteString("\n\n[" + img.Name + "](" + img.URL + ")")
	}
	assistant.Parts[t.textPart] = message.TextContent{Text: sb.String()}
	t.images.claim(images)
	if err := t.msgs.Update(context.Background(), *assistant); err != nil {
		slog.Warn("Failed to store unlinked cloud images", "error", err)
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
		// The canonical ACP image block: the agent sends the picture
		// itself, and its prose refers to it with Markdown image
		// syntax. Keep the bytes locally and link the reference.
		if img := update.AgentMessageChunk.Content.Image; img != nil {
			return t.applyImage(img)
		}
		text := contentBlockText(update.AgentMessageChunk.Content)
		if text == "" {
			return nil
		}
		return t.appendText(text)

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
	t.saveToolImages(tu.Content)
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
