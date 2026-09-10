package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/teams"
)

//go:embed team.md
var teamDescription string

const TeamToolName = "team"

type TeamParams struct {
	Action      string   `json:"action" description:"Action: join, leave, list_agents, create_task, list_tasks, claim_task, update_task, send_message, or inbox"`
	Team        string   `json:"team" description:"Team name"`
	Agent       string   `json:"agent,omitempty" description:"Agent name; defaults to the current session's registered agent"`
	Role        string   `json:"role,omitempty" description:"Agent role, such as architect, coder, reviewer, researcher, or tester"`
	Subject     string   `json:"subject,omitempty" description:"Short task subject"`
	Description string   `json:"description,omitempty" description:"Detailed task description"`
	TaskID      string   `json:"task_id,omitempty" description:"Task identifier"`
	Status      string   `json:"status,omitempty" description:"Task status: pending, claimed, running, blocked, completed, or failed"`
	Result      string   `json:"result,omitempty" description:"Task result or failure details"`
	BlockedBy   []string `json:"blocked_by,omitempty" description:"Task IDs that must complete before this task can be claimed"`
	To          string   `json:"to,omitempty" description:"Recipient agent name, or all for broadcast"`
	Message     string   `json:"message,omitempty" description:"Message content"`
	UnreadOnly  bool     `json:"unread_only,omitempty" description:"Only return unread inbox messages"`
	MarkRead    bool     `json:"mark_read,omitempty" description:"Mark returned inbox messages as read"`
}

func NewTeamTool(store *teams.Store) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		TeamToolName,
		teamDescription,
		func(ctx context.Context, params TeamParams, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if strings.TrimSpace(params.Team) == "" {
				return fantasy.NewTextErrorResponse("team is required"), nil
			}
			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, errors.New("session ID is required for team coordination")
			}
			agentName := strings.TrimSpace(params.Agent)
			if agentName == "" {
				agentName = sessionID
			}

			var (
				value any
				err   error
			)
			switch params.Action {
			case "join":
				role := strings.TrimSpace(params.Role)
				if role == "" {
					role = "coder"
				}
				value, err = store.Join(ctx, params.Team, teams.Agent{Name: agentName, Role: role, SessionID: sessionID, Status: teams.AgentIdle})
			case "leave":
				value, err = store.Leave(ctx, params.Team, agentName, teams.AgentStopped)
			case "list_agents":
				value, err = store.ListAgents(ctx, params.Team)
			case "create_task":
				if strings.TrimSpace(params.Subject) == "" {
					return fantasy.NewTextErrorResponse("subject is required to create a task"), nil
				}
				value, err = store.CreateTask(ctx, params.Team, params.Subject, params.Description, params.BlockedBy)
			case "list_tasks":
				value, err = store.ListTasks(ctx, params.Team)
			case "claim_task":
				if params.TaskID == "" {
					return fantasy.NewTextErrorResponse("task_id is required to claim a task"), nil
				}
				value, err = store.ClaimTask(ctx, params.Team, params.TaskID, agentName)
			case "update_task":
				if params.TaskID == "" || params.Status == "" {
					return fantasy.NewTextErrorResponse("task_id and status are required to update a task"), nil
				}
				value, err = store.UpdateTask(ctx, params.Team, params.TaskID, agentName, teams.TaskStatus(params.Status), params.Result)
			case "send_message":
				if params.To == "" || strings.TrimSpace(params.Message) == "" {
					return fantasy.NewTextErrorResponse("to and message are required to send a message"), nil
				}
				value, err = store.SendMessage(ctx, params.Team, teams.Message{From: agentName, To: params.To, Content: params.Message})
			case "inbox":
				value, err = store.Inbox(ctx, params.Team, agentName, params.UnreadOnly, params.MarkRead)
			default:
				return fantasy.NewTextErrorResponse("invalid action"), nil
			}
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			output, err := json.MarshalIndent(value, "", "  ")
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("encode team response: %w", err)
			}
			return fantasy.NewTextResponse(string(output)), nil
		},
	)
}
