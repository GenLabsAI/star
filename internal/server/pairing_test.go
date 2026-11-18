package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func pairedTestServer(token string) http.Handler {
	srv := &Server{network: "tcp", Token: token}
	return srv.pairingHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("reached the api"))
	}))
}

func doRequest(handler http.Handler, header, value string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	if header != "" {
		req.Header.Set(header, value)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestUnpairedRequestIsRejected(t *testing.T) {
	t.Parallel()

	handler := pairedTestServer("secret-token")
	response := doRequest(handler, "", "")
	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.Contains(t, response.Body.String(), "pairing token")
}

func TestWrongTokenIsRejected(t *testing.T) {
	t.Parallel()

	handler := pairedTestServer("secret-token")
	response := doRequest(handler, "Authorization", "Bearer wrong-token")
	require.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestPairedRequestReachesTheAPI(t *testing.T) {
	t.Parallel()

	handler := pairedTestServer("secret-token")

	response := doRequest(handler, "Authorization", "Bearer secret-token")
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "reached the api", response.Body.String())

	// A peer may prefer its own header.
	response = doRequest(handler, "X-Star-Token", "secret-token")
	require.Equal(t, http.StatusOK, response.Code)
}

func TestNoTokenConfiguredIsUnchanged(t *testing.T) {
	t.Parallel()

	handler := pairedTestServer("")
	response := doRequest(handler, "", "")
	require.Equal(t, http.StatusOK, response.Code)
}
