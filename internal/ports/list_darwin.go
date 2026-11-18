//go:build darwin

package ports

import (
	"context"
	"fmt"
	"os/exec"
)

// list asks lsof for listening TCP sockets. lsof exits non-zero when it
// finds nothing, which is an empty inventory rather than a failure.
// supportsCWD reports whether this platform attributes sockets to a
// process working directory.
const supportsCWD = false

func list(ctx context.Context) ([]Listener, error) {
	out, err := exec.CommandContext(ctx, "lsof", "-nP", "-iTCP", "-sTCP:LISTEN", "-F", "pcn").Output()
	if err != nil && len(out) == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to list listening sockets: %w", err)
	}
	return Dedupe(parseLsof(string(out))), nil
}
