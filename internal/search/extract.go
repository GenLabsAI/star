package search

import (
	"strings"

	"github.com/charmbracelet/crush/internal/message"
)

// ExtractSearchableText walks a message's parts and returns a single
// concatenated string containing all text suitable for full-text search
// indexing.
func ExtractSearchableText(parts []message.ContentPart) string {
	var text strings.Builder
	for _, p := range parts {
		switch part := p.(type) {
		case message.TextContent:
			text.WriteString(part.Text)
			text.WriteByte('\n')
		case message.ReasoningContent:
			text.WriteString(part.Thinking)
			text.WriteByte('\n')
		case message.ToolCall:
			text.WriteString(part.Name)
			text.WriteByte('\n')
		case message.ToolResult:
			text.WriteString(part.Content)
			text.WriteByte('\n')
		case message.ShellCommand:
			text.WriteString(part.Command)
			text.WriteByte('\n')
			text.WriteString(part.Output)
			text.WriteByte('\n')
		}
	}
	return text.String()
}
