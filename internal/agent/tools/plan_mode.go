package tools

import (
	"context"
	_ "embed"
	"fmt"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/permission"
)

//go:embed enter_plan_mode.md
var enterPlanModeDescription string

//go:embed exit_plan_mode.md
var exitPlanModeDescription string

const (
	EnterPlanModeToolName = "enter_plan_mode"
	ExitPlanModeToolName  = "exit_plan_mode"
)

type PlanModeParams struct{}

func newSetPlanModeTool(permissions permission.Service, name, description, response string, mode permission.Mode) fantasy.AgentTool {
	return fantasy.NewAgentTool(name, description, func(ctx context.Context, params PlanModeParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
		sessionID := GetSessionFromContext(ctx)
		if sessionID == "" {
			return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for %s", name)
		}
		permissions.SetSessionMode(sessionID, mode)
		return fantasy.NewTextResponse(response), nil
	})
}

func NewEnterPlanModeTool(permissions permission.Service) fantasy.AgentTool {
	return newSetPlanModeTool(permissions, EnterPlanModeToolName, enterPlanModeDescription, "Entered Plan Mode.", permission.ModePlan)
}

func NewExitPlanModeTool(permissions permission.Service) fantasy.AgentTool {
	return newSetPlanModeTool(permissions, ExitPlanModeToolName, exitPlanModeDescription, "Exited Plan Mode and entered Normal Mode.", permission.ModeNormal)
}
