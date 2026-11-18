// Package ports inventories the TCP sockets listening on the machine,
// with the process and working directory behind each one, so an agent can
// find the dev server it started (or someone else did) without guessing
// ports or scraping ps output.
package ports

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// Listener is one listening TCP socket.
type Listener struct {
	// Port is the listening port.
	Port int `json:"port"`
	// Address is the bind address, e.g. 0.0.0.0, 127.0.0.1 or ::.
	Address string `json:"address"`
	// PID is the owning process id, zero when it could not be resolved.
	PID int `json:"pid,omitempty"`
	// Process is the process name, empty when it could not be resolved.
	Process string `json:"process,omitempty"`
	// CWD is the process working directory, empty when it could not be
	// resolved.
	CWD string `json:"cwd,omitempty"`
	// Workspace reports whether CWD is inside the requested workspace.
	Workspace bool `json:"workspace"`
}

// List returns every listening TCP socket the process can see.
func List(ctx context.Context) ([]Listener, error) {
	return list(ctx)
}

// Within marks the listeners whose process working directory is inside
// dir, and returns them.
func Within(listeners []Listener, dir string) []Listener {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil
	}
	prefix := abs + string(filepath.Separator)
	within := make([]Listener, 0, len(listeners))
	for _, listener := range listeners {
		if listener.CWD == "" {
			continue
		}
		if listener.CWD == abs || strings.HasPrefix(listener.CWD, prefix) {
			listener.Workspace = true
			within = append(within, listener)
		}
	}
	return within
}

// SupportsCWD reports whether this platform's collector can attribute a
// listener to a process working directory. Where it cannot, filtering by
// workspace would hide everything, so callers should show every listener.
func SupportsCWD() bool {
	return supportsCWD
}

// Dedupe drops repeated (port, pid) rows, which happen when IPv4 and IPv6
// sockets are reported separately for the same process.
func Dedupe(listeners []Listener) []Listener {
	seen := make(map[string]struct{}, len(listeners))
	unique := make([]Listener, 0, len(listeners))
	for _, listener := range listeners {
		key := fmt.Sprintf("%d|%s|%d", listener.Port, listener.Address, listener.PID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, listener)
	}
	return unique
}
