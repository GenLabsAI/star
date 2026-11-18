package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/crush/internal/pairing"
	"github.com/stretchr/testify/require"
)

// policyTestServer wraps a handler that always answers 200, behind the
// pairing check and the policy check, so a test sees exactly which gate a
// request stopped at.
func policyTestServer(token string, policy PairPolicy) http.Handler {
	srv := &Server{network: "tcp", Token: token}
	srv.SetPairPolicy(policy)
	return srv.pairingHandler(srv.pairPolicyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("reached the api"))
	})))
}

func policyRequest(handler http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	var req *http.Request
	if reader != nil {
		req = httptest.NewRequest(method, path, reader)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestReadOnlyTokenCannotWrite(t *testing.T) {
	t.Parallel()

	token, err := pairing.Mint(pairing.ScopeSet{pairing.ScopeRead})
	require.NoError(t, err)
	handler := policyTestServer(token, PairPolicy{})
	auth := map[string]string{"Authorization": "Bearer " + token}

	response := policyRequest(handler, http.MethodGet, "/v1/workspaces", "", auth)
	require.Equal(t, http.StatusOK, response.Code)

	response = policyRequest(handler, http.MethodPost, "/v1/workspaces", `{"path":"/tmp/x"}`, auth)
	require.Equal(t, http.StatusForbidden, response.Code)
	require.Contains(t, response.Body.String(), "this pairing token grants")
	require.Contains(t, response.Body.String(), "requires")
	require.Contains(t, response.Body.String(), "write")

	// The exec route is refused too, and before any approval question.
	response = policyRequest(handler, http.MethodPost, "/v1/workspaces/id/agent/sessions/sid/shell", "{}", auth)
	require.Equal(t, http.StatusForbidden, response.Code)
	require.Contains(t, response.Body.String(), "exec")
}

func TestExecNeedsACurrentApproval(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "approval.json")
	minted, err := pairing.SaveApproval(path, time.Minute)
	require.NoError(t, err)
	token, err := pairing.Mint(pairing.AllScopes())
	require.NoError(t, err)
	handler := policyTestServer(token, PairPolicy{ApprovalFile: path})
	auth := map[string]string{"Authorization": "Bearer " + token}
	execPath := "/v1/workspaces/id/agent/sessions/sid/shell"

	response := policyRequest(handler, http.MethodPost, execPath, "{}", auth)
	require.Equal(t, http.StatusPreconditionRequired, response.Code)
	require.Contains(t, response.Body.String(), "approval required")

	auth["X-Star-Approval"] = "WRONGCODE"
	response = policyRequest(handler, http.MethodPost, execPath, "{}", auth)
	require.Equal(t, http.StatusPreconditionRequired, response.Code)

	auth[pairing.ApprovalHeader] = minted.Code
	response = policyRequest(handler, http.MethodPost, execPath, "{}", auth)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "reached the api", response.Body.String())

	// Reading still needs no approval: the gate is on running things.
	delete(auth, pairing.ApprovalHeader)
	response = policyRequest(handler, http.MethodGet, "/v1/workspaces", "", auth)
	require.Equal(t, http.StatusOK, response.Code)
}

func TestExpiredApprovalIsRefused(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "approval.json")
	minted, err := pairing.SaveApproval(path, time.Millisecond)
	require.NoError(t, err)
	time.Sleep(5 * time.Millisecond)
	token, err := pairing.Mint(pairing.AllScopes())
	require.NoError(t, err)
	handler := policyTestServer(token, PairPolicy{ApprovalFile: path})

	response := policyRequest(handler, http.MethodPost,
		"/v1/workspaces/id/agent/sessions/sid/shell", "{}",
		map[string]string{
			"Authorization":        "Bearer " + token,
			pairing.ApprovalHeader: minted.Code,
		})
	require.Equal(t, http.StatusPreconditionRequired, response.Code)
	require.Contains(t, response.Body.String(), "expired")
}

func TestWorkspaceRestrictionOnCreate(t *testing.T) {
	t.Parallel()

	allowed := t.TempDir()
	token, err := pairing.Mint(pairing.AllScopes())
	require.NoError(t, err)
	handler := policyTestServer(token, PairPolicy{Workspaces: []string{allowed}})
	auth := map[string]string{"Authorization": "Bearer " + token}

	response := policyRequest(handler, http.MethodPost, "/v1/workspaces",
		`{"path":"/somewhere/else"}`, auth)
	require.Equal(t, http.StatusForbidden, response.Code)
	require.Contains(t, response.Body.String(), "may only use the workspace")

	response = policyRequest(handler, http.MethodPost, "/v1/workspaces",
		`{"path":"`+filepath.Join(allowed, "inside")+`"}`, auth)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "reached the api", response.Body.String())

	// Requests that do not name a workspace are untouched.
	response = policyRequest(handler, http.MethodGet, "/v1/health", "", auth)
	require.Equal(t, http.StatusOK, response.Code)
}

func TestUnrestrictedTokenKeepsWorking(t *testing.T) {
	t.Parallel()

	token, err := pairing.Mint(pairing.AllScopes())
	require.NoError(t, err)
	handler := policyTestServer(token, PairPolicy{})
	response := policyRequest(handler, http.MethodPost, "/v1/workspaces", `{"path":"/tmp/x"}`,
		map[string]string{"Authorization": "Bearer " + token})
	require.Equal(t, http.StatusOK, response.Code)
}
