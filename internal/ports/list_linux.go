//go:build linux

package ports

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// list reads the kernel's socket tables directly, so it works in minimal
// containers with no ss, lsof or netstat installed. ss is the fallback for
// kernels or sandboxes that hide /proc/net.
// supportsCWD reports whether this platform attributes sockets to a
// process working directory.
const supportsCWD = true

func list(ctx context.Context) ([]Listener, error) {
	if listeners, err := procfsListeners(); err == nil {
		return Dedupe(listeners), nil
	}
	return ssListeners(ctx)
}

func procfsListeners() ([]Listener, error) {
	data, err := os.ReadFile("/proc/net/tcp")
	if err != nil {
		return nil, fmt.Errorf("Failed to read /proc/net/tcp: %w", err)
	}
	sockets := parseProcNet(string(data))
	if six, err := os.ReadFile("/proc/net/tcp6"); err == nil {
		sockets = append(sockets, parseProcNet(string(six))...)
	}
	if len(sockets) == 0 {
		return nil, nil
	}

	owners := socketOwners()
	listeners := make([]Listener, 0, len(sockets))
	for _, socket := range sockets {
		listener := Listener{Port: socket.Port, Address: socket.Address}
		if pid, ok := owners[socket.Inode]; ok {
			listener.PID = pid
			listener.Process = readComm(pid)
			listener.CWD = readCWD(pid)
		}
		listeners = append(listeners, listener)
	}
	return listeners, nil
}

// socketOwners maps socket inodes to the owning process, by walking every
// process's file descriptors. Cheap enough for an inventory command.
func socketOwners() map[string]int {
	owners := map[string]int{}
	processes, err := os.ReadDir("/proc")
	if err != nil {
		return owners
	}
	for _, process := range processes {
		if !process.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(process.Name())
		if err != nil {
			continue
		}
		fdDir := filepath.Join("/proc", process.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, fd := range fds {
			target, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			inode, ok := strings.CutPrefix(target, "socket:[")
			if !ok {
				continue
			}
			owners[strings.TrimSuffix(inode, "]")] = pid
		}
	}
	return owners
}

func readComm(pid int) string {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "comm"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readCWD(pid int) string {
	cwd, err := os.Readlink(filepath.Join("/proc", strconv.Itoa(pid), "cwd"))
	if err != nil {
		return ""
	}
	return cwd
}

func ssListeners(ctx context.Context) ([]Listener, error) {
	out, err := exec.CommandContext(ctx, "ss", "-ltnp").Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to list listening sockets: %w", err)
	}
	return Dedupe(parseSS(string(out))), nil
}
