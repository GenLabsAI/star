// Enforcement of what a paired peer may do. The pairing token answers who
// the peer is; this file answers what it is allowed to touch: a scope per
// request, an optional approval code in front of running commands, and an
// optional list of workspaces the peer may use.
package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/pairing"
)

// maxCreateBodyBytes bounds the part of a workspace-create body read to
// learn its path. A larger body is not a create request worth honouring.
const maxCreateBodyBytes = 1 << 20

// PairPolicy is what a paired peer may do beyond proving it holds the
// token. The token's scopes come with the token; this policy adds the
// server-side limits an operator chooses.
type PairPolicy struct {
	// Workspaces limits the peer to these directories. Empty means every
	// workspace.
	Workspaces []string
	// ApprovalFile makes running commands require a current approval code
	// minted there. Empty means no approval is needed.
	ApprovalFile string
}

// SetPairPolicy applies the limits on what a paired peer may do.
func (s *Server) SetPairPolicy(policy PairPolicy) {
	s.PairWorkspaces = policy.Workspaces
	s.ApprovalFile = policy.ApprovalFile
}

// requiredScope reports what a request needs. Reading is GET and HEAD,
// running anything is the shell route, and everything else changes state.
func requiredScope(method, path string) pairing.Scope {
	switch {
	case strings.Contains(path, "/shell"), strings.Contains(path, "/terminal"):
		return pairing.ScopeExec
	case method == http.MethodGet, method == http.MethodHead, method == http.MethodOptions:
		return pairing.ScopeRead
	default:
		return pairing.ScopeWrite
	}
}

// workspaceIDFromPath returns the workspace id a path names, if any.
func workspaceIDFromPath(path string) string {
	rest, ok := strings.CutPrefix(path, "/v1/workspaces/")
	if !ok {
		return ""
	}
	id, _, _ := strings.Cut(rest, "/")
	return id
}

// createWorkspacePath returns the path a workspace-create request asks for,
// leaving the body readable for the handler.
func createWorkspacePath(r *http.Request) string {
	if r.Method != http.MethodPost || strings.TrimSuffix(r.URL.Path, "/") != "/v1/workspaces" {
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxCreateBodyBytes))
	if err != nil {
		return ""
	}
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(body, &args); err != nil {
		return ""
	}
	return args.Path
}

// workspaceAllowed reports whether a directory is inside one of the
// workspaces this pairing may use. No list means every workspace.
func (s *Server) workspaceAllowed(path string) bool {
	if len(s.PairWorkspaces) == 0 || path == "" {
		return true
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, dir := range s.PairWorkspaces {
		root, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if abs == root || strings.HasPrefix(abs, root+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// presentedApproval reads the approval code from its header.
func presentedApproval(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get(pairing.ApprovalHeader))
}

// pairPolicyHandler applies the token's scopes, the approval required for
// running commands, and the workspace restriction. It runs after the
// pairing check, so every request here already proved it holds the token.
//
// The policy is read per request because it is configured after the server
// exists, and because an approval file can be minted or expire while the
// server keeps running.
func (s *Server) pairPolicyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.Token == "" {
			next.ServeHTTP(w, r)
			return
		}

		scopes := pairing.ScopesOf(s.Token)
		need := requiredScope(r.Method, r.URL.Path)
		if !scopes.Grants(need) {
			s.logError(r, "Rejected request outside the token's scopes",
				"path", r.URL.Path, "needed", string(need), "granted", scopes.String())
			jsonError(w, http.StatusForbidden, fmt.Sprintf(
				"forbidden: this pairing token grants %q and %s %s requires %q",
				scopes.String(), r.Method, r.URL.Path, string(need)))
			return
		}

		if need == pairing.ScopeExec && s.ApprovalFile != "" {
			approval, err := pairing.ReadApproval(s.ApprovalFile)
			switch {
			case err != nil:
				s.logError(r, "Rejected exec without an approval file", "error", err)
				jsonError(w, http.StatusPreconditionRequired,
					"approval required: this server runs commands only with an approval code (mint one with `star pair --approval --out FILE`)")
				return
			case !time.Now().Before(approval.ExpiresAt):
				s.logError(r, "Rejected exec with an expired approval")
				jsonError(w, http.StatusPreconditionRequired,
					"approval required: the approval code has expired (mint a new one with `star pair --approval`)")
				return
			case !approval.Valid(presentedApproval(r), time.Now()):
				s.logError(r, "Rejected exec without a valid approval code")
				jsonError(w, http.StatusPreconditionRequired,
					"approval required: present a current approval code in "+pairing.ApprovalHeader)
				return
			}
		}

		if len(s.PairWorkspaces) > 0 {
			if path := createWorkspacePath(r); path != "" && !s.workspaceAllowed(path) {
				s.logError(r, "Rejected workspace outside the granted ones", "path", path)
				jsonError(w, http.StatusForbidden, fmt.Sprintf(
					"forbidden: this pairing may only use the workspace(s) %s, not %s",
					strings.Join(s.PairWorkspaces, ", "), path))
				return
			}
			// An id the server does not know falls through to the
			// handler's 404: refusing it here would leak which ids
			// exist on this server.
			if id := workspaceIDFromPath(r.URL.Path); id != "" && s.backend != nil {
				if ws, err := s.backend.GetWorkspace(id); err == nil && !s.workspaceAllowed(ws.Path) {
					s.logError(r, "Rejected request for a workspace outside the granted ones", "id", id, "path", ws.Path)
					jsonError(w, http.StatusForbidden, fmt.Sprintf(
						"forbidden: this pairing may only use the workspace(s) %s",
						strings.Join(s.PairWorkspaces, ", ")))
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
