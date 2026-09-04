package backend

import (
	"github.com/charmbracelet/crush/internal/permission"
	"github.com/charmbracelet/crush/internal/proto"
)

// GrantPermission grants, denies, or persistently grants a permission
// request. The returned bool reports whether this call resolved the
// pending request (true) or found it already resolved by a previous
// caller (false). A false return is not an error.
func (b *Backend) GrantPermission(workspaceID string, req proto.PermissionGrant) (bool, error) {
	ws, err := b.GetWorkspace(workspaceID)
	if err != nil {
		return false, err
	}

	perm := permission.PermissionRequest{
		ID:          req.Permission.ID,
		SessionID:   req.Permission.SessionID,
		ToolCallID:  req.Permission.ToolCallID,
		ToolName:    req.Permission.ToolName,
		Description: req.Permission.Description,
		Action:      req.Permission.Action,
		Params:      req.Permission.Params,
		Path:        req.Permission.Path,
	}

	switch req.Action {
	case proto.PermissionAllow:
		return ws.Permissions.Grant(perm), nil
	case proto.PermissionAllowForSession:
		return ws.Permissions.GrantPersistent(perm), nil
	case proto.PermissionDeny:
		return ws.Permissions.Deny(perm), nil
	default:
		return false, ErrInvalidPermissionAction
	}
}

// SetPermissionsSkip sets whether permission prompts are skipped. When
// sessionID is non-empty, the mode is scoped to that session only;
// otherwise it sets the workspace-wide default.
func (b *Backend) SetPermissionsSkip(workspaceID, sessionID string, skip bool) error {
	ws, err := b.GetWorkspace(workspaceID)
	if err != nil {
		return err
	}

	if sessionID != "" {
		ws.Permissions.SetSessionSkipRequests(sessionID, skip)
		return nil
	}
	ws.Permissions.SetSkipRequests(skip)
	return nil
}

// GetPermissionsSkip returns whether permission prompts are skipped for
// the given session (or the workspace-wide default when sessionID is
// empty).
func (b *Backend) GetPermissionsSkip(workspaceID, sessionID string) (bool, error) {
	ws, err := b.GetWorkspace(workspaceID)
	if err != nil {
		return false, err
	}

	if sessionID != "" {
		return ws.Permissions.SessionSkipRequests(sessionID), nil
	}
	return ws.Permissions.SkipRequests(), nil
}

// SetPermissionsMode sets the permission mode. When sessionID is
// non-empty, the mode is scoped to that session only; otherwise it sets
// the workspace-wide default.
func (b *Backend) SetPermissionsMode(workspaceID, sessionID string, mode permission.Mode) error {
	ws, err := b.GetWorkspace(workspaceID)
	if err != nil {
		return err
	}

	if sessionID != "" {
		ws.Permissions.SetSessionMode(sessionID, mode)
		return nil
	}
	ws.Permissions.SetMode(mode)
	return nil
}

// GetPermissionsMode returns the permission mode for the given session
// (or the workspace-wide default when sessionID is empty).
func (b *Backend) GetPermissionsMode(workspaceID, sessionID string) (permission.Mode, error) {
	ws, err := b.GetWorkspace(workspaceID)
	if err != nil {
		return "", err
	}

	if sessionID != "" {
		return ws.Permissions.SessionMode(sessionID), nil
	}
	return ws.Permissions.GetMode(), nil
}
