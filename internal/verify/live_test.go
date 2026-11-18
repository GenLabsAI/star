package verify

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseStatusList(t *testing.T) {
	t.Parallel()

	codes, err := ParseStatusList("")
	require.NoError(t, err)
	require.Equal(t, []int{200}, codes)

	codes, err = ParseStatusList("200, 204,200")
	require.NoError(t, err)
	require.Equal(t, []int{200, 204}, codes)

	_, err = ParseStatusList("200,abc")
	require.ErrorContains(t, err, "Invalid status")

	_, err = ParseStatusList("99")
	require.ErrorContains(t, err, "between 100 and 599")

	_, err = ParseStatusList("200,,204")
	require.ErrorContains(t, err, "empty entry")
}

func TestGateLive(t *testing.T) {
	t.Parallel()

	passing := &Live{Passed: true}
	failing := &Live{Passed: false}

	require.Equal(t, VerdictPass, gateLive(VerdictPass, nil))
	require.Equal(t, VerdictPass, gateLive(VerdictPass, passing))
	require.Equal(t, VerdictPass, gateLive(VerdictNoCommands, passing))
	require.Equal(t, VerdictLiveFail, gateLive(VerdictPass, failing))
	require.Equal(t, VerdictLiveFail, gateLive(VerdictNoCommands, failing))
	// A failing or flaky test run keeps its own verdict: the live phase
	// only decides a run that would otherwise pass.
	require.Equal(t, VerdictFail, gateLive(VerdictFail, failing))
	require.Equal(t, VerdictFlaky, gateLive(VerdictFlaky, failing))
}

func TestRunLiveFindsTheApplication(t *testing.T) {
	// Not parallel: the helper reads the environment the test sets.
	port := freePort(t)
	t.Setenv(helperEnv, "1")
	t.Setenv(helperPortEnv, strconv.Itoa(port))
	t.Setenv(helperStatusEnv, "200")

	bundle := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(bundle, logDirName), 0o755))

	live := runLive(context.Background(), t.TempDir(), liveSpec{
		command:      fmt.Sprintf("%q -test.run=TestLiveHelperProcess", os.Args[0]),
		url:          fmt.Sprintf("http://127.0.0.1:%d/health", port),
		expect:       []int{200},
		readyTimeout: 30 * time.Second,
	}, bundle, 2000)

	require.True(t, live.Passed, "reason: %s", live.Reason)
	require.True(t, live.Ready)
	require.Equal(t, 200, live.StatusCode)
	require.Contains(t, live.BodyTail, "ok")
	require.Contains(t, live.OutputTail, "helper serving")

	log, err := os.ReadFile(filepath.Join(bundle, live.Log))
	require.NoError(t, err)
	require.Contains(t, string(log), "# status: 200")
}

func TestRunLiveReportsTheWrongStatus(t *testing.T) {
	port := freePort(t)
	t.Setenv(helperEnv, "1")
	t.Setenv(helperPortEnv, strconv.Itoa(port))
	t.Setenv(helperStatusEnv, "500")

	bundle := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(bundle, logDirName), 0o755))

	live := runLive(context.Background(), t.TempDir(), liveSpec{
		command:      fmt.Sprintf("%q -test.run=TestLiveHelperProcess", os.Args[0]),
		url:          fmt.Sprintf("http://127.0.0.1:%d/health", port),
		expect:       []int{200},
		readyTimeout: 30 * time.Second,
	}, bundle, 2000)

	require.False(t, live.Passed)
	require.Equal(t, 500, live.StatusCode)
	require.Contains(t, live.Reason, "answered 500, expected 200")
	require.Contains(t, live.OutputTail, "helper serving")
}

func TestRunLiveReportsAnApplicationThatExits(t *testing.T) {
	bundle := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(bundle, logDirName), 0o755))

	live := runLive(context.Background(), t.TempDir(), liveSpec{
		command:      "echo 'cannot bind'; exit 3",
		url:          "http://127.0.0.1:1/health",
		expect:       []int{200},
		readyTimeout: 5 * time.Second,
	}, bundle, 2000)

	require.False(t, live.Passed)
	require.False(t, live.Ready)
	require.Contains(t, live.Reason, "exited before answering")
	require.Contains(t, live.Reason, "exit status 3")
	require.Contains(t, live.OutputTail, "cannot bind")
}

const (
	helperEnv       = "STAR_LIVE_TEST_HELPER"
	helperPortEnv   = "STAR_LIVE_TEST_PORT"
	helperStatusEnv = "STAR_LIVE_TEST_STATUS"
)

// TestLiveHelperProcess is not a test: it is the application the live
// phase starts, which is this test binary re-executed. It only runs when
// the environment names it.
func TestLiveHelperProcess(t *testing.T) {
	if os.Getenv(helperEnv) == "" {
		t.Skip("not the live helper")
	}
	port := os.Getenv(helperPortEnv)
	status, err := strconv.Atoi(os.Getenv(helperStatusEnv))
	if err != nil {
		status = http.StatusOK
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte("ok"))
	})
	fmt.Println("helper serving on " + port)
	if err := http.ListenAndServe("127.0.0.1:"+port, mux); err != nil {
		fmt.Println("helper failed: " + err.Error())
		os.Exit(1)
	}
}

// freePort returns a port that was free a moment ago.
func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	return port
}
