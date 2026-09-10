package worktree

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// Manager handles isolated git worktrees for sessions/agents.
type Manager interface {
	Create(ctx context.Context, sessionID, branch string) (string, error)
	Remove(ctx context.Context, sessionID string) error
	Cleanup(ctx context.Context) error
}

type manager struct {
	basePath string
	rootDir  string
	mu       sync.Mutex
	paths    map[string]string
}

// NewManager creates a worktree manager rooted in the given repository.
func NewManager(rootDir, basePath string) Manager {
	return &manager{basePath: basePath, rootDir: rootDir, paths: make(map[string]string)}
}

func (m *manager) Create(ctx context.Context, sessionID, branch string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.paths[sessionID]; ok {
		return existing, nil
	}
	if err := os.MkdirAll(m.basePath, 0o755); err != nil {
		return "", fmt.Errorf("create worktree directory: %w", err)
	}
	path := filepath.Join(m.basePath, sessionID)
	args := []string{"worktree", "add", "--detach", path}
	if branch != "" {
		args = []string{"worktree", "add", "-b", branch, path}
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = m.rootDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("create git worktree: %w: %s", err, output)
	}
	m.paths[sessionID] = path
	return path, nil
}

func (m *manager) Remove(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	path, ok := m.paths[sessionID]
	if ok {
		delete(m.paths, sessionID)
	}
	m.mu.Unlock()
	if !ok {
		return nil
	}
	return m.remove(ctx, path)
}

func (m *manager) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	paths := make([]string, 0, len(m.paths))
	for _, path := range m.paths {
		paths = append(paths, path)
	}
	m.paths = make(map[string]string)
	m.mu.Unlock()
	var lastErr error
	for _, path := range paths {
		if err := m.remove(ctx, path); err != nil {
			slog.Warn("Failed to remove git worktree during cleanup", "path", path, "error", err)
			lastErr = err
		}
	}
	return lastErr
}

func (m *manager) remove(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "git", "worktree", "remove", "--force", path)
	cmd.Dir = m.rootDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("remove git worktree: %w: %s", err, output)
	}
	cmd = exec.CommandContext(ctx, "git", "worktree", "prune")
	cmd.Dir = m.rootDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("prune git worktrees: %w: %s", err, output)
	}
	return nil
}
