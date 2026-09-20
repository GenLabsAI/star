package search

import (
	"github.com/charmbracelet/crush/internal/message"
)

// ExtractSearchableText walks a message's parts and returns a single concatenated
// string containing all text suitable for full-text search indexing.
func ExtractSearchableText(parts []message.ContentPart) string {
	var text string
	for _, p := range parts {
		switch part := p.(type) {
		case message.TextContent:
			text += part.Text + "\n"
		case message.ReasoningContent:
			text += part.Thinking + "\n"
		case message.ToolCall:
			text += part.Name + "\n"
		case message.ToolResult:
			text += part.Content + "\n"
		case message.ShellCommand:
			text += part.Command + "\n" + part.Output + "\n"
		}
	}
	return text
}
