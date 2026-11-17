package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/crush/internal/lock"
	"github.com/google/uuid"
)

// WakeupSchedule is one durable wakeup: a timed fire (one-shot,
// repeating, or cron), an armed named event, or both. Schedules
// survive restarts because they live in the workspace data directory
// instead of in process memory.
type WakeupSchedule struct {
	ID            string `json:"id"`
	SessionID     string `json:"session_id"`
	Reason        string `json:"reason"`
	Prompt        string `json:"prompt"`
	Event         string `json:"event,omitempty"`
	Cloud         bool   `json:"cloud"`
	RepeatSeconds int64  `json:"repeat_seconds,omitempty"`
	Cron          string `json:"cron,omitempty"`
	NextFireUnix  int64  `json:"next_fire_unix,omitempty"`
	CreatedUnix   int64  `json:"created_unix"`
}

// Timed reports whether the schedule has a timer arm (as opposed to
// an event-only schedule that waits for FireEvent).
func (s WakeupSchedule) Timed() bool {
	return s.NextFireUnix > 0
}

// Repeating reports whether the schedule recurs after firing.
func (s WakeupSchedule) Repeating() bool {
	return s.RepeatSeconds > 0 || s.Cron != ""
}

// Repeat describes the schedule's repetition for display.
func (s WakeupSchedule) Repeat() string {
	switch {
	case s.Cron != "":
		return "cron " + s.Cron
	case s.RepeatSeconds > 0:
		return "every " + (time.Duration(s.RepeatSeconds) * time.Second).String()
	default:
		return "once"
	}
}

// WakeupStore persists schedules across restarts. Implementations
// must be safe for concurrent use within a process; the file store
// additionally serializes across processes with an advisory lock.
type WakeupStore interface {
	Save(ctx context.Context, schedule WakeupSchedule) error
	Delete(ctx context.Context, id string) error
	DeleteSession(ctx context.Context, sessionID string) (int, error)
	List(ctx context.Context) ([]WakeupSchedule, error)
	// ClaimDue atomically takes ownership of one due timed schedule:
	// one-shot schedules are removed, repeating ones advance to
	// their next fire. It reports false when another process claimed
	// the schedule first.
	ClaimDue(ctx context.Context, id string, now time.Time) (WakeupSchedule, bool, error)
	// Consume removes one schedule by id, reporting whether it
	// existed. Event firings consume one-shots atomically so two
	// processes racing the same event cannot both run it.
	Consume(ctx context.Context, id string) (bool, error)
}

// NewWakeupStore opens the durable store for a workspace data
// directory. An empty dataDir (or any I/O failure) falls back to an
// in-memory store so tests and exotic setups keep working.
func NewWakeupStore(dataDir string) WakeupStore {
	if dataDir == "" {
		return NewMemoryWakeupStore()
	}
	path := filepath.Join(dataDir, "wakeups", "schedules.json")
	return &fileWakeupStore{path: path, lockPath: path + ".lock"}
}

// NewWakeupScheduleID mints a schedule id.
func NewWakeupScheduleID() string {
	return "wake-" + uuid.New().String()
}

type wakeupFile struct {
	Version   int              `json:"version"`
	Schedules []WakeupSchedule `json:"schedules"`
}

// memoryWakeupStore is the non-durable fallback.
type memoryWakeupStore struct {
	mu        sync.Mutex
	schedules map[string]WakeupSchedule
}

// NewMemoryWakeupStore returns an in-memory WakeupStore.
func NewMemoryWakeupStore() WakeupStore {
	return &memoryWakeupStore{schedules: make(map[string]WakeupSchedule)}
}

func (s *memoryWakeupStore) Save(_ context.Context, schedule WakeupSchedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules[schedule.ID] = schedule
	return nil
}

func (s *memoryWakeupStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.schedules, id)
	return nil
}

func (s *memoryWakeupStore) Consume(_ context.Context, id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.schedules[id]; !ok {
		return false, nil
	}
	delete(s.schedules, id)
	return true, nil
}

func (s *memoryWakeupStore) DeleteSession(_ context.Context, sessionID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for id, sched := range s.schedules {
		if sched.SessionID == sessionID {
			delete(s.schedules, id)
			n++
		}
	}
	return n, nil
}

func (s *memoryWakeupStore) List(_ context.Context) ([]WakeupSchedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]WakeupSchedule, 0, len(s.schedules))
	for _, sched := range s.schedules {
		out = append(out, sched)
	}
	return out, nil
}

func (s *memoryWakeupStore) ClaimDue(_ context.Context, id string, now time.Time) (WakeupSchedule, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sched, ok := s.schedules[id]
	if !ok || !sched.Timed() || sched.NextFireUnix > now.Unix() {
		return WakeupSchedule{}, false, nil
	}
	if next, ok := advanceSchedule(sched, now); ok {
		sched.NextFireUnix = next.Unix()
		s.schedules[id] = sched
	} else {
		delete(s.schedules, id)
	}
	return sched, true, nil
}

// fileWakeupStore persists schedules as JSON under the workspace
// data directory. Every mutation takes an inter-process lock, reads
// the file fresh, mutates, and writes atomically, so a TUI, a
// server, and a cron-invoked CLI sharing one data directory cannot
// lose or double-fire each other's schedules.
type fileWakeupStore struct {
	mu       sync.Mutex
	path     string
	lockPath string
}

func (s *fileWakeupStore) withFile(ctx context.Context, fn func(*wakeupFile) (bool, error)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("Failed to create wakeup store directory: %w", err)
	}
	lockCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	release, err := lock.File(lockCtx, s.lockPath)
	if err != nil {
		return fmt.Errorf("Failed to lock wakeup store: %w", err)
	}
	defer release()
	data, err := readWakeupFile(s.path)
	if err != nil {
		return err
	}
	dirty, err := fn(data)
	if err != nil {
		return err
	}
	if !dirty {
		return nil
	}
	return writeWakeupFile(s.path, data)
}

func (s *fileWakeupStore) Save(ctx context.Context, schedule WakeupSchedule) error {
	if schedule.ID == "" {
		return fmt.Errorf("Wakeup schedule needs an ID")
	}
	return s.withFile(ctx, func(data *wakeupFile) (bool, error) {
		for i, existing := range data.Schedules {
			if existing.ID == schedule.ID {
				data.Schedules[i] = schedule
				return true, nil
			}
		}
		data.Schedules = append(data.Schedules, schedule)
		return true, nil
	})
}

func (s *fileWakeupStore) Delete(ctx context.Context, id string) error {
	removed := false
	return s.withFile(ctx, func(data *wakeupFile) (bool, error) {
		kept := data.Schedules[:0]
		for _, sched := range data.Schedules {
			if sched.ID == id {
				removed = true
				continue
			}
			kept = append(kept, sched)
		}
		data.Schedules = kept
		return removed, nil
	})
}

func (s *fileWakeupStore) Consume(ctx context.Context, id string) (bool, error) {
	consumed := false
	err := s.withFile(ctx, func(data *wakeupFile) (bool, error) {
		kept := data.Schedules[:0]
		for _, sched := range data.Schedules {
			if sched.ID == id {
				consumed = true
				continue
			}
			kept = append(kept, sched)
		}
		data.Schedules = kept
		return consumed, nil
	})
	return consumed, err
}

func (s *fileWakeupStore) DeleteSession(ctx context.Context, sessionID string) (int, error) {
	n := 0
	err := s.withFile(ctx, func(data *wakeupFile) (bool, error) {
		kept := data.Schedules[:0]
		for _, sched := range data.Schedules {
			if sched.SessionID == sessionID {
				n++
				continue
			}
			kept = append(kept, sched)
		}
		data.Schedules = kept
		return n > 0, nil
	})
	return n, err
}

func (s *fileWakeupStore) List(ctx context.Context) ([]WakeupSchedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := readWakeupFile(s.path)
	if err != nil {
		return nil, err
	}
	out := make([]WakeupSchedule, len(data.Schedules))
	copy(out, data.Schedules)
	return out, nil
}

func (s *fileWakeupStore) ClaimDue(ctx context.Context, id string, now time.Time) (WakeupSchedule, bool, error) {
	var claimed WakeupSchedule
	found := false
	err := s.withFile(ctx, func(data *wakeupFile) (bool, error) {
		for i, sched := range data.Schedules {
			if sched.ID != id {
				continue
			}
			if !sched.Timed() || sched.NextFireUnix > now.Unix() {
				return false, nil
			}
			claimed = sched
			found = true
			if next, ok := advanceSchedule(sched, now); ok {
				data.Schedules[i].NextFireUnix = next.Unix()
			} else {
				data.Schedules = append(data.Schedules[:i], data.Schedules[i+1:]...)
			}
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return WakeupSchedule{}, false, err
	}
	return claimed, found, nil
}

func readWakeupFile(path string) (*wakeupFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &wakeupFile{Version: 1}, nil
		}
		return nil, fmt.Errorf("Failed to read wakeup store: %w", err)
	}
	var data wakeupFile
	if err := json.Unmarshal(raw, &data); err != nil {
		slog.Warn("Ignoring corrupt wakeup store; starting fresh", "path", path, "error", err)
		return &wakeupFile{Version: 1}, nil
	}
	return &data, nil
}

func writeWakeupFile(path string, data *wakeupFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("Failed to create wakeup store directory: %w", err)
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to encode wakeup store: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "schedules-*.tmp")
	if err != nil {
		return fmt.Errorf("Failed to write wakeup store: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("Failed to write wakeup store: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("Failed to write wakeup store: %w", err)
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("Failed to secure wakeup store: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("Failed to write wakeup store: %w", err)
	}
	return nil
}

// advanceSchedule computes the next fire after now for a repeating
// schedule. It reports false for one-shot schedules, which the
// caller removes instead of advancing.
func advanceSchedule(sched WakeupSchedule, now time.Time) (time.Time, bool) {
	switch {
	case sched.Cron != "":
		next, err := NextCronOccurrence(sched.Cron, now)
		if err != nil {
			slog.Warn("Dropping schedule with invalid cron", "schedule", sched.ID, "cron", sched.Cron)
			return time.Time{}, false
		}
		return next, true
	case sched.RepeatSeconds > 0:
		next := time.Unix(sched.NextFireUnix, 0)
		step := time.Duration(sched.RepeatSeconds) * time.Second
		for !next.After(now) {
			next = next.Add(step)
		}
		return next, true
	default:
		return time.Time{}, false
	}
}
