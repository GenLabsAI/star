package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/charmbracelet/crush/internal/pairing"
	"github.com/stretchr/testify/require"
)

// TestClientPresentsPairingToken proves a peer configured with the pairing
// token sends it on every request, which is what makes an authorized
// client work against a paired server.
func TestClientPresentsPairingToken(t *testing.T) {
	t.Setenv(pairing.EnvToken, "pair-token")
	t.Setenv(pairing.EnvTokenFile, "")

	var authorization string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := NewClient(t.TempDir(), "tcp", strings.TrimPrefix(srv.URL, "http://"))
	require.NoError(t, err)
	require.NoError(t, c.Health(context.Background()))
	require.Equal(t, "Bearer pair-token", authorization)
}

func TestClientWithoutTokenSendsNoAuthorization(t *testing.T) {
	t.Setenv(pairing.EnvToken, "")
	t.Setenv(pairing.EnvTokenFile, "")

	var authorization string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := NewClient(t.TempDir(), "tcp", strings.TrimPrefix(srv.URL, "http://"))
	require.NoError(t, err)
	require.NoError(t, c.Health(context.Background()))
	require.Empty(t, authorization)
}
