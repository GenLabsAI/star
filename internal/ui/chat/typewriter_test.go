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

	started, _ := item.BufferStreamingMessage(updated)
	require.True(t, started)
	require.True(t, item.HasBufferedContent())

	item.AdvanceBufferedContent()
	require.Equal(t, "A", item.displayedContent)

	item.AdvanceBufferedContent()
	require.Equal(t, "A⭐", item.displayedContent)
	require.True(t, utf8.ValidString(item.displayedContent))

	item.AdvanceBufferedContent()
	require.Equal(t, "A⭐B", item.displayedContent)
	require.False(t, item.HasBufferedContent())
}

func TestAssistantMessageTypewriter_FinishWithBacklog(t *testing.T) {
	sty := styles.CharmtonePantera()
	msg := &message.Message{ID: "a", Role: message.Assistant}
	item := NewAssistantMessageItem(&sty, msg).(*AssistantMessageItem)

	// Provide target text, simulate streaming updates
	update := &message.Message{ID: "a", Role: message.Assistant, Parts: []message.ContentPart{message.TextContent{Text: "hello"}}}
	item.BufferStreamingMessage(update)

	// Should not be finished initially because of spinning/content
	require.False(t, item.Finished())

	// Finish the message while buffer is not fully consumed
	updateFinished := &message.Message{
		ID: "a", Role: message.Assistant,
		Parts: []message.ContentPart{
			message.TextContent{Text: "hello"},
			message.Finish{Reason: message.FinishReasonEndTurn},
		},
	}
	item.BufferStreamingMessage(updateFinished)

	// Even though message is finished, we still have buffered content
	require.True(t, item.HasBufferedContent())
	require.False(t, item.Finished(), "Finished() should wait for buffered content to drain")

	// Drain buffer
	for item.HasBufferedContent() {
		item.AdvanceBufferedContent()
	}

	// Now it is finished
	require.False(t, item.HasBufferedContent())
	require.True(t, item.Finished())
}

func TestAssistantMessageTypewriter_RewriteResetsBuffer(t *testing.T) {
	sty := styles.CharmtonePantera()
	msg := &message.Message{ID: "a", Role: message.Assistant}
	item := NewAssistantMessageItem(&sty, msg).(*AssistantMessageItem)

	// Provide some streaming content
	item.BufferStreamingMessage(&message.Message{ID: "a", Role: message.Assistant, Parts: []message.ContentPart{message.TextContent{Text: "hello"}}})
	item.AdvanceBufferedContent() // "h"
	item.AdvanceBufferedContent() // "he"
	require.Equal(t, "he", item.displayedContent)

	// Simulate a rewrite (e.g. from a tool hook or a retry)
	// The new content is not prefixed by "he"
	rewrite := &message.Message{ID: "a", Role: message.Assistant, Parts: []message.ContentPart{message.TextContent{Text: "bonjour"}}}
	item.BufferStreamingMessage(rewrite)

	// Since the prefix didn't match, it should have reset displayed to the new content to prevent visual corruption
	require.Equal(t, "bonjour", item.displayedContent)
	require.False(t, item.HasBufferedContent())
}
