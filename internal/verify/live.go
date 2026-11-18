// Live verification: start the application itself, probe it over HTTP and
// keep its log, so a failure is reproduced against the running process
// rather than its test suite.
//
// Nothing here trusts the process to be healthy because it started. A run
// is only a pass when the application answered, and answered with a status
// the caller asked for; anything else is reported with the answer, the
// reason and the tail of the application's own log.
package verify

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/crush/internal/shell"
)

// VerdictLiveFail means the application ran but did not answer the way the
// caller required: it never came up, it exited first, or it answered with
// an unexpected status.
const VerdictLiveFail = "live-fail"

// Defaults for the live phase.
const (
	// DefaultReadyTimeout is how long an application may take to answer
	// its first request.
	DefaultReadyTimeout = 30 * time.Second
	liveLogName         = "90-app.log"
	// maxBodyBytes bounds the part of the response body kept as evidence.
	maxBodyBytes = 4096
	// maxAppLogBytes bounds the application log kept in memory. The end
	// of the log is what explains a failure, so the oldest bytes go.
	maxAppLogBytes = 256 << 10
	// stopGrace bounds how long a stopped application may take to exit.
	stopGrace = 5 * time.Second
	// probeInterval is the pause between readiness probes.
	probeInterval = 150 * time.Millisecond
)

// Live is what happened when the application ran: whether it came up, what
// it answered, and what it logged while doing it.
type Live struct {
	// Command is the command that started the application.
	Command string `json:"command"`
	// URL is the address that was probed.
	URL string `json:"url"`
	// Ready reports whether the application answered at all.
	Ready bool `json:"ready"`
	// ReadyMS is how long it took to answer, when it did.
	ReadyMS int64 `json:"ready_ms,omitempty"`
	// StatusCode is the status it answered with, zero when it never did.
	StatusCode int `json:"status_code,omitempty"`
	// Expected lists the statuses that count as healthy.
	Expected []int `json:"expected_status"`
	// Passed reports whether it answered with one of them.
	Passed bool `json:"passed"`
	// Reason explains a failure in the application's own terms.
	Reason string `json:"reason,omitempty"`
	// BodyTail is the end of the response body that was inspected.
	BodyTail string `json:"body_tail,omitempty"`
	// OutputTail is the end of the application log.
	OutputTail string `json:"output_tail,omitempty"`
	// Log is the application log file, relative to the bundle.
	Log string `json:"log,omitempty"`
}

// liveSpec is the resolved form of the live options.
type liveSpec struct {
	command      string
	url          string
	expect       []int
	readyTimeout time.Duration
}

// ParseStatusList turns "200,204" into the statuses a healthy answer may
// have. An empty list means the single status 200.
func ParseStatusList(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []int{http.StatusOK}, nil
	}
	fields := strings.Split(raw, ",")
	codes := make([]int, 0, len(fields))
	seen := make(map[int]struct{}, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			return nil, fmt.Errorf("Invalid status list %q: empty entry", raw)
		}
		code, err := strconv.Atoi(field)
		if err != nil || code < 100 || code > 599 {
			return nil, fmt.Errorf("Invalid status %q: want an HTTP status between 100 and 599", field)
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	return codes, nil
}

// statusList renders the expected statuses for a message.
func statusList(codes []int) string {
	parts := make([]string, 0, len(codes))
	for _, code := range codes {
		parts = append(parts, strconv.Itoa(code))
	}
	return strings.Join(parts, " or ")
}

// gateLive folds the live phase into the verdict. A failing or flaky test
// run stays what it is; the live phase decides only a run that would
// otherwise pass, and it is the whole result when there were no commands
// to run at all.
func gateLive(verdict string, live *Live) string {
	if live == nil {
		return verdict
	}
	if live.Passed {
		if verdict == VerdictNoCommands {
			return VerdictPass
		}
		return verdict
	}
	switch verdict {
	case VerdictPass, VerdictNoCommands:
		return VerdictLiveFail
	}
	return verdict
}

// runLive starts the application, waits for it to answer, and stops it
// again. It always returns a record: a failure to start is evidence, not
// an error to swallow.
func runLive(ctx context.Context, dir string, spec liveSpec, bundle string, tailBytes int) *Live {
	live := &Live{Command: spec.command, URL: spec.url, Expected: spec.expect}
	live.Log = filepath.ToSlash(filepath.Join(logDirName, liveLogName))

	out := &appLog{}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	started := time.Now()
	go func() {
		done <- shell.Run(runCtx, shell.RunOptions{
			Command: spec.command,
			Cwd:     dir,
			Env:     os.Environ(),
			Stdout:  out,
			Stderr:  out,
		})
	}()

	timeout := spec.readyTimeout
	if timeout <= 0 {
		timeout = DefaultReadyTimeout
	}
	probe := &http.Client{
		Timeout: 5 * time.Second,
		// A proxy configured for the machine must not answer for the
		// application on loopback, and a redirect is an answer worth
		// reporting rather than following.
		Transport: &http.Transport{Proxy: nil, DisableKeepAlives: true},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	deadline := started.Add(timeout)
	for {
		request, err := http.NewRequestWithContext(runCtx, http.MethodGet, spec.url, nil)
		var response *http.Response
		if err == nil {
			response, err = probe.Do(request)
		}
		if err == nil {
			live.Ready = true
			live.ReadyMS = time.Since(started).Milliseconds()
			live.StatusCode = response.StatusCode
			body, _ := io.ReadAll(io.LimitReader(response.Body, maxBodyBytes))
			_ = response.Body.Close()
			live.BodyTail = strings.TrimSpace(string(body))
			break
		}

		// A process that exited is not going to answer later; saying so
		// now beats waiting out the whole timeout.
		select {
		case runErr := <-done:
			live.Reason = fmt.Sprintf("the application exited before answering %s (%s)", spec.url, exitText(runErr))
			live.OutputTail = tail(strings.TrimSpace(out.String()), tailBytes)
			writeLiveLog(bundle, live, out.String())
			return live
		case <-time.After(probeInterval):
		}

		if time.Now().After(deadline) {
			live.Reason = fmt.Sprintf("the application did not answer %s within %s: %v", spec.url, timeout, err)
			live.OutputTail = tail(strings.TrimSpace(out.String()), tailBytes)
			break
		}
	}

	// Stop the application and let it go, so a run leaves nothing behind.
	cancel()
	select {
	case <-done:
	case <-time.After(stopGrace):
		live.Reason = joinReason(live.Reason, fmt.Sprintf("the application did not stop within %s and was killed", stopGrace))
	}

	live.Passed = live.Ready && containsStatus(spec.expect, live.StatusCode)
	if !live.Passed && live.Reason == "" {
		if live.Ready {
			live.Reason = fmt.Sprintf("the application answered %d, expected %s", live.StatusCode, statusList(spec.expect))
		} else {
			live.Reason = fmt.Sprintf("the application never answered %s", spec.url)
		}
	}
	if live.OutputTail == "" {
		live.OutputTail = tail(strings.TrimSpace(out.String()), tailBytes)
	}
	writeLiveLog(bundle, live, out.String())
	return live
}

// writeLiveLog keeps the application's own output next to the record, with
// the command and outcome a reader needs to judge it.
func writeLiveLog(bundle string, live *Live, output string) {
	if live.Log == "" {
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "$ %s\n", live.Command)
	fmt.Fprintf(&b, "# url: %s\n", live.URL)
	fmt.Fprintf(&b, "# ready: %t\n", live.Ready)
	fmt.Fprintf(&b, "# status: %d\n", live.StatusCode)
	if live.Reason != "" {
		fmt.Fprintf(&b, "# reason: %s\n", live.Reason)
	}
	b.WriteString("\n")
	b.WriteString(strings.TrimRight(output, "\n"))
	b.WriteString("\n")
	_ = os.WriteFile(filepath.Join(bundle, filepath.FromSlash(live.Log)), []byte(b.String()), 0o644)
}

// joinReason appends a second explanation to the first.
func joinReason(first, second string) string {
	if first == "" {
		return second
	}
	return first + "; " + second
}

// exitText describes how the application ended before it answered.
func exitText(err error) string {
	if err == nil {
		return "exit status 0"
	}
	return err.Error()
}

// bodyLine keeps a summary line short by showing only the body's first
// line.
func bodyLine(body string) string {
	if line, _, ok := strings.Cut(strings.TrimSpace(body), "\n"); ok {
		return line
	}
	return strings.TrimSpace(body)
}

func containsStatus(codes []int, code int) bool {
	for _, candidate := range codes {
		if candidate == code {
			return true
		}
	}
	return false
}

// appLog keeps the last maxAppLogBytes the application wrote, so a runaway
// logger cannot exhaust memory while the end of the log, where the failure
// is, survives.
type appLog struct {
	mu  sync.Mutex
	buf []byte
}

func (l *appLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = append(l.buf, p...)
	if over := len(l.buf) - maxAppLogBytes; over > 0 {
		l.buf = l.buf[over:]
	}
	return len(p), nil
}

func (l *appLog) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return string(l.buf)
}
