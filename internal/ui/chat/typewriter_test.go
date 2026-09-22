package chat

import (
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/stretchr/testify/require"
)

func TestAssistantMessageTypewriterBuffersStreamingContent(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	msg := &message.Message{ID: "assistant", Role: message.Assistant}
	item := NewAssistantMessageItem(&sty, msg).(*AssistantMessageItem)

	updated := &message.Message{
		ID:   "assistant",
		Role: message.Assistant,
		Parts: []message.ContentPart{
			message.TextContent{Text: "A⭐B"},
		},
	}

	item.SetMessage(updated)
	item.displayedContent = ""
	item.BufferContent(updated.Content().Text)
	require.True(t, item.HasBufferedContent())
	require.Empty(t, item.displayedContent)

	item.AdvanceBufferedContent()
	require.Equal(t, "A", item.displayedContent)

	item.AdvanceBufferedContent()
	require.Equal(t, "A⭐", item.displayedContent)
	require.True(t, utf8.ValidString(item.displayedContent))

	item.AdvanceBufferedContent()
	require.Equal(t, "A⭐B", item.displayedContent)
	require.False(t, item.HasBufferedContent())
}
