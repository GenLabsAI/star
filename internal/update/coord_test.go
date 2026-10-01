package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func withCoordDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("STAR_UPDATE_DIR", dir)
	return dir
}

func TestRequestPendingIgnoresRequestsFromBeforeLaunch(t *testing.T) {
	dir := withCoordDir(t)

	if err := PublishRequest(); err != nil {
		t.Fatalf("PublishRequest: %v", err)
	}

	// Backdate the request to simulate one left behind by an update that
	// crashed, and then treat "now" as this instance's launch time. A stale
	// request must not tear down an instance that just booted.
	stale := time.Now().Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(dir, requestFile), stale, stale); err != nil {
		t.Fatalf("backdating request: %v", err)
	}

	if RequestPending(time.Now().UnixNano()) {
		t.Fatal("a request older than this launch must not pause the instance")
	}
}

func TestRequestPendingSeesRequestsFromAfterLaunch(t *testing.T) {
	withCoordDir(t)

	launched := time.Now().Add(-time.Minute).UnixNano()
	if err := PublishRequest(); err != nil {
		t.Fatalf("PublishRequest: %v", err)
	}
	// Filesystem timestamp granularity varies, so give the request a
	// comfortably newer mtime than the launch stamp being compared against.

	if !RequestPending(launched) {
		t.Fatal("a request made after this launch must pause the instance")
	}
}

func TestRequestPendingWithoutRequestFile(t *testing.T) {
	withCoordDir(t)

	if RequestPending(time.Now().Add(-time.Hour).UnixNano()) {
		t.Fatal("no request file means nothing to pause for")
	}
}

func TestPublishRequestCreatesCoordDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "not", "there", "yet")
	t.Setenv("STAR_UPDATE_DIR", dir)

	if err := PublishRequest(); err != nil {
		t.Fatalf("PublishRequest: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, requestFile)); err != nil {
		t.Fatalf("request file was not created: %v", err)
	}
}

func TestRecordSessionRoundTrip(t *testing.T) {
	dir := withCoordDir(t)

	if err := RecordSession("1234", "sess-abc"); err != nil {
		t.Fatalf("RecordSession: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, sessionsDir, "1234"))
	if err != nil {
		t.Fatalf("reading session: %v", err)
	}
	if string(got) != "sess-abc" {
		t.Fatalf("got session %q, want %q", got, "sess-abc")
	}
}

func TestRecordSessionIgnoresEmptyValues(t *testing.T) {
	dir := withCoordDir(t)

	if err := RecordSession("1234", ""); err != nil {
		t.Fatalf("RecordSession: %v", err)
	}
	if err := RecordSession("", "sess-abc"); err != nil {
		t.Fatalf("RecordSession: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, sessionsDir)); !os.IsNotExist(err) {
		t.Fatal("nothing should have been written")
	}
}
