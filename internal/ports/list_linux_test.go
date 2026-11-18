//go:build linux

package ports

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestListFindsRealListener binds a real socket and checks the inventory
// reports it with this process's pid and working directory, which is what
// `star ports` relies on.
func TestListFindsRealListener(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	listeners, err := List(context.Background())
	require.NoError(t, err)

	var found *Listener
	for i := range listeners {
		if listeners[i].Port == port {
			found = &listeners[i]
			break
		}
	}
	require.NotNil(t, found, "listener on port %d not found", port)
	require.Equal(t, os.Getpid(), found.PID)
	require.Equal(t, "127.0.0.1", found.Address)
	require.NotEmpty(t, found.CWD)

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, wd, found.CWD)
}
