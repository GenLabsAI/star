package acp

import (
	"log/slog"

	"charm.land/catwalk/pkg/catwalk"
	"github.com/charmbracelet/crush/internal/agent/hyper"
	"github.com/charmbracelet/crush/internal/agent/tools/mcp"
	"github.com/charmbracelet/crush/internal/commands"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/coder/acp-go-sdk"
)

const mcpCommandPrefix = "mcp:"

// HandleMCPEvent processes MCP events and republishes commands when prompts
// change.
func (s *Sink) HandleMCPEvent(event pubsub.Event[mcp.Event]) {
	switch event.Payload.Type {
	case mcp.EventPromptsListChanged, mcp.EventStateChanged:
		s.PublishCommands()
	}
}

// PublishCommands aggregates commands from all sources and sends an
// AvailableCommandsUpdate to the ACP client.
func (s *Sink) PublishCommands() {
	var availableCommands []acp.AvailableCommand

	// System/built-in commands.
	availableCommands = append(availableCommands, s.builtinCommands()...)

	// User and project commands (already prefixed by commands).
	if userCmds, err := commands.LoadCustomCommands(s.configStore.Config()); err == nil {
		availableCommands = append(availableCommands, translateCommands(userCmds, "",
			func(c commands.CustomCommand) string { return c.ID },
			func(c commands.CustomCommand) string {
				if c.Skill != nil {
					return c.Skill.Description
				}
				return ""
			},
			func(c commands.CustomCommand) string { return c.Name })...)
	}

	// MCP prompts.
	if mcpCmds, err := commands.LoadMCPPrompts(); err == nil {
		availableCommands = append(availableCommands, translateCommands(mcpCmds, mcpCommandPrefix,
			func(c commands.MCPPrompt) string { return c.ID },
			func(c commands.MCPPrompt) string { return c.Description },
			func(c commands.MCPPrompt) string { return c.Title })...)
	}

	if err := s.conn.SessionUpdate(s.ctx, acp.SessionNotification{
		SessionId: acp.SessionId(s.sessionID),
		Update: acp.SessionUpdate{
			AvailableCommandsUpdate: &acp.SessionAvailableCommandsUpdate{
				AvailableCommands: availableCommands,
			},
		},
	}); err != nil {
		slog.Error("Failed to send available commands update", "error", err)
	}
}

// builtinCommands returns ACP-compatible built-in commands.
// Commands are dynamically generated based on current model capabilities.
func (s *Sink) builtinCommands() []acp.AvailableCommand {
	commands := []acp.AvailableCommand{
		{
			Name:        "summarize",
			Description: "Summarize the current session and create a new one with the summary",
		},
		{
			Name:        "toggle_yolo",
			Description: "Toggle yolo mode (auto-approve tool calls)",
		},
	}

	// Add reasoning commands based on current model capabilities.
	if s.configStore == nil {
		return commands
	}
	cfg := s.configStore.Config()
	if cfg == nil {
		return commands
	}

	agentCfg, ok := cfg.Agents[config.AgentCoder]
	if !ok {
		return commands
	}

	providerCfg := cfg.GetProviderForModel(agentCfg.Model)
	model := cfg.GetModelByType(agentCfg.Model)
	if providerCfg == nil || model == nil || !model.CanReason {
		return commands
	}

	// Anthropic/Hyper models: thinking toggle.
	if providerCfg.Type == catwalk.TypeAnthropic || providerCfg.Type == catwalk.Type(hyper.Name) {
		commands = append(commands, acp.AvailableCommand{
			Name:        "toggle_thinking",
			Description: "Toggle extended thinking for reasoning-capable models",
		})
	}

	// OpenAI-style models: reasoning effort selection.
	if len(model.ReasoningLevels) > 0 {
		commands = append(commands, acp.AvailableCommand{
			Name:        "set_reasoning_effort",
			Description: "Set reasoning effort level (low, medium, high)",
			Input: &acp.AvailableCommandInput{
				UnstructuredCommandInput: &acp.AvailableCommandUnstructuredCommandInput{
					Hint: "low | medium | high",
				},
			},
		})
	}

	return commands
}

// translateCommands converts command slices to acp.AvailableCommand slice,
// optionally adding a prefix.
//
// Star's command structs expose plain fields rather than the getter
// methods the upstream Crush PR expects, so callers pass field
// accessors instead.
func translateCommands[T any](cmds []T, prefix string, getID func(T) string, getDescription func(T) string, getTitle func(T) string) []acp.AvailableCommand {
	result := make([]acp.AvailableCommand, 0, len(cmds))
	for _, cmd := range cmds {
		id := getID(cmd)
		title := getTitle(cmd)
		acpCmd := acp.AvailableCommand{
			Name:        prefix + id,
			Description: getDescription(cmd),
		}

		// If the command has a title different from ID, use it as a hint.
		if title != "" && title != id {
			acpCmd.Input = &acp.AvailableCommandInput{
				UnstructuredCommandInput: &acp.AvailableCommandUnstructuredCommandInput{
					Hint: title,
				},
			}
		}

		result = append(result, acpCmd)
	}
	return result
}
