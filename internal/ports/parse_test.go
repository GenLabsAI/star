package ports

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeProcAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		field string
		addr  string
		port  int
		ok    bool
	}{
		{"0100007F:1F90", "127.0.0.1", 8080, true},
		{"00000000:0CEA", "0.0.0.0", 3306, true},
		{"00000000000000000000000000000000:0050", "::", 80, true},
		{"00000000000000000000000001000000:0016", "::1", 22, true},
		{"0000000000000000ffff000000000000:01BB", "0.0.0.0", 443, true},
		{"nope", "", 0, false},
		{"0100007F:ZZZZ", "", 0, false},
	}
	for _, test := range tests {
		addr, port, ok := decodeProcAddress(test.field)
		require.Equal(t, test.ok, ok, test.field)
		if !test.ok {
			continue
		}
		require.Equal(t, test.addr, addr, test.field)
		require.Equal(t, test.port, port, test.field)
	}
}

func TestParseProcNetOnlyKeepsListeners(t *testing.T) {
	t.Parallel()

	data := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 11111 1 0
   1: 0100007F:1F91 00000000:0000 01 00000000:00000000 00:00000000 00000000  1000        0 22222 1 0
`
	sockets := parseProcNet(data)
	require.Len(t, sockets, 1)
	require.Equal(t, procSocket{Port: 8080, Address: "127.0.0.1", Inode: "11111"}, sockets[0])
}

func TestParseSS(t *testing.T) {
	t.Parallel()

	data := `State  Recv-Q Send-Q Local Address:Port  Peer Address:Port Process
LISTEN 0      4096         0.0.0.0:8787       0.0.0.0:*     users:(("node",pid=4242,fd=21))
LISTEN 0      128        127.0.0.1:5432       0.0.0.0:*     users:(("postgres",pid=99,fd=7))
ESTAB  0      0        127.0.0.1:22        127.0.0.1:5555
`
	listeners := parseSS(data)
	require.Len(t, listeners, 2)
	require.Equal(t, Listener{Port: 8787, Address: "0.0.0.0", PID: 4242, Process: "node"}, listeners[0])
	require.Equal(t, Listener{Port: 5432, Address: "127.0.0.1", PID: 99, Process: "postgres"}, listeners[1])
}

func TestParseSSWithIPv6(t *testing.T) {
	t.Parallel()

	data := "LISTEN 0 128 [::]:8080 [::]:* users:((\"python3\",pid=7,fd=3))\n"
	listeners := parseSS(data)
	require.Len(t, listeners, 1)
	require.Equal(t, "::", listeners[0].Address)
	require.Equal(t, 8080, listeners[0].Port)
	require.Equal(t, 7, listeners[0].PID)
}

func TestParseLsof(t *testing.T) {
	t.Parallel()

	data := "p4242\ncnode\nn*:3000\nn127.0.0.1:9229\np77\ncpostgres\nn[::1]:5432\n"
	listeners := parseLsof(data)
	require.Equal(t, []Listener{
		{Port: 3000, Address: "*", PID: 4242, Process: "node"},
		{Port: 9229, Address: "127.0.0.1", PID: 4242, Process: "node"},
		{Port: 5432, Address: "::1", PID: 77, Process: "postgres"},
	}, listeners)
}

func TestParseNetstat(t *testing.T) {
	t.Parallel()

	data := `
  Proto  Local Address          Foreign Address        State           PID
  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       944
  TCP    127.0.0.1:3000         0.0.0.0:0              LISTENING       1234
  TCP    127.0.0.1:5000         127.0.0.1:61000        ESTABLISHED     1234
`
	listeners := parseNetstat(data)
	require.Equal(t, []Listener{
		{Port: 135, Address: "0.0.0.0", PID: 944},
		{Port: 3000, Address: "127.0.0.1", PID: 1234},
	}, listeners)
}

func TestWithinMarksWorkspaceListeners(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	listeners := []Listener{
		{Port: 1, CWD: dir},
		{Port: 2, CWD: dir + "/sub"},
		{Port: 3, CWD: "/somewhere/else"},
		{Port: 4},
	}
	within := Within(listeners, dir)
	require.Len(t, within, 2)
	require.Equal(t, 1, within[0].Port)
	require.True(t, within[0].Workspace)
	require.Equal(t, 2, within[1].Port)
}

func TestDedupe(t *testing.T) {
	t.Parallel()

	listeners := []Listener{
		{Port: 80, Address: "0.0.0.0", PID: 1, CWD: "/x"},
		{Port: 80, Address: "0.0.0.0", PID: 1, CWD: "/x"},
		{Port: 80, Address: "0.0.0.0", PID: 2},
	}
	require.Len(t, Dedupe(listeners), 2)
}
