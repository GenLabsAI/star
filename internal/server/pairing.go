package server

import (
	"net/http"
	"strings"

	"github.com/charmbracelet/crush/internal/pairing"
)

// SetPairToken makes the server require the token on every request. An
// empty token leaves the server in its default local-only mode, where the
// socket permissions are the boundary.
func (s *Server) SetPairToken(token string) {
	s.Token = token
}

// pairingHandler rejects requests that do not present the configured
// pairing token. It runs before the API, so an unpaired peer cannot read
// workspaces, sessions or files, and cannot run anything.
//
// The token is read per request, not captured when the handler chain is
// built, because the token is configured after the server exists.
func (s *Server) pairingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.Token == "" {
			next.ServeHTTP(w, r)
			return
		}
		if !pairing.Match(presentedToken(r), s.Token) {
			s.logError(r, "Rejected unpaired request", "path", r.URL.Path)
			jsonError(w, http.StatusUnauthorized,
				"unauthorized: this server requires a pairing token (see `star pair`)")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// presentedToken reads the peer's token from the Authorization header, or
// from X-Star-Token for clients that prefer a header of their own.
func presentedToken(r *http.Request) string {
	if token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return strings.TrimSpace(token)
	}
	return strings.TrimSpace(r.Header.Get("X-Star-Token"))
}
