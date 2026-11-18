//go:build windows

package ports

import (
	"context"
	"fmt"
	"os/exec"
)

// list asks netstat for listening TCP sockets. Windows exposes the owning
// pid but not its working directory from here.
// supportsCWD reports whether this platform attributes sockets to a
// process working directory.
const supportsCWD = false

func list(ctx context.Context) ([]Listener, error) {
	out, err := exec.CommandContext(ctx, "netstat", "-ano", "-p", "tcp").Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to list listening sockets: %w", err)
	}
	return Dedupe(parseNetstat(string(out))), nil
}
