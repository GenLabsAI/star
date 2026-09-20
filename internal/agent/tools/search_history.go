package tools

import (
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/search"
)

type SearchHistoryParams struct {
	Query     string `json:"query" description:"Search query to find relevant past conversations and context"`
	SessionID string `json:"session_id,omitempty" description:"Optional: Limit search to a specific session ID"`
	Limit     int    `json:"limit,omitempty" description:"Max results to return (default 10, max 50)"`
}

// NewSearchHistoryTool creates a tool that lets the agent search its own conversation history.
func NewSearchHistoryTool(svc search.Service) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"search_history",
		"Search your conversation history across past sessions to recall previous decisions, context, code patterns, or solutions. Use this when the user references a past conversation, says 'remember when we...', or when you need context about the project that was discussed before.",
		func(ctx context.Context, params SearchHistoryParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			limit := params.Limit
			if limit <= 0 {
				limit = 10
			} else if limit > 50 {
				limit = 50
			}

			opts := search.SearchOpts{
				Limit: limit,
			}
			if params.SessionID != "" {
				opts.SessionID = params.SessionID
			}

			results, err := svc.Search(ctx, params.Query, opts)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to search history: %v", err)), nil
			}

			if len(results) == 0 {
				return fantasy.NewTextResponse("No relevant conversations found in history."), nil
			}

			var b strings.Builder
			fmt.Fprintf(&b, "Found %d results for '%s':\n\n", len(results), params.Query)

			for i, res := range results {
				// Remove ANSI highlights if present
				cleanSnippet := strings.ReplaceAll(res.Snippet, "{{", "")
				cleanSnippet = strings.ReplaceAll(cleanSnippet, "}}", "")

				fmt.Fprintf(&b, "[%d] Session: %s (ID: %s)\n", i+1, res.SessionTitle, res.SessionID)
				fmt.Fprintf(&b, "Role: %s\n", res.Role)
				fmt.Fprintf(&b, "Snippet: %s\n\n", cleanSnippet)
			}

			return fantasy.NewTextResponse(b.String()), nil
		},
	)
}
