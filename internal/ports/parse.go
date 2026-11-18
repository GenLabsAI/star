package ports

import (
	"encoding/hex"
	"net"
	"strconv"
	"strings"
)

// procSocket is one listening socket from /proc/net/tcp{,6}.
type procSocket struct {
	Port    int
	Address string
	Inode   string
}

// parseProcNet decodes /proc/net/tcp or /proc/net/tcp6. The local address
// column is hex-encoded little-endian; only rows in state 0A (LISTEN)
// count.
func parseProcNet(data string) []procSocket {
	var sockets []procSocket
	for i, line := range strings.Split(data, "\n") {
		if i == 0 {
			continue // Header.
		}
		fields := strings.Fields(line)
		if len(fields) < 10 || fields[3] != "0A" {
			continue
		}
		address, port, ok := decodeProcAddress(fields[1])
		if !ok {
			continue
		}
		sockets = append(sockets, procSocket{Port: port, Address: address, Inode: fields[9]})
	}
	return sockets
}

// decodeProcAddress decodes "0100007F:1F90" into 127.0.0.1 and 8080. The
// kernel prints the address as 32-bit words in host (little-endian)
// order, for IPv6 too, so each four-byte group is reversed.
func decodeProcAddress(field string) (string, int, bool) {
	hexAddr, hexPort, ok := strings.Cut(field, ":")
	if !ok {
		return "", 0, false
	}
	port, err := strconv.ParseUint(hexPort, 16, 16)
	if err != nil {
		return "", 0, false
	}
	raw, err := hex.DecodeString(hexAddr)
	if err != nil || len(raw)%4 != 0 {
		return "", 0, false
	}
	for i := 0; i < len(raw); i += 4 {
		raw[i], raw[i+3] = raw[i+3], raw[i]
		raw[i+1], raw[i+2] = raw[i+2], raw[i+1]
	}
	ip := net.IP(raw)
	if ip.To4() != nil {
		return ip.To4().String(), int(port), true
	}
	return ip.String(), int(port), true
}

// parseSS decodes `ss -ltnp` output, the fallback when /proc is
// unavailable.
func parseSS(data string) []Listener {
	var listeners []Listener
	for _, line := range strings.Split(data, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "LISTEN") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		address, port := splitHostPort(fields[3])
		listener := Listener{Port: port, Address: address}
		for _, field := range fields[4:] {
			pid, process, ok := parseSSUsers(field)
			if ok {
				listener.PID = pid
				listener.Process = process
			}
		}
		if listener.Port != 0 {
			listeners = append(listeners, listener)
		}
	}
	return listeners
}

// parseSSUsers decodes `users:(("python3",pid=1234,fd=3))`.
func parseSSUsers(field string) (int, string, bool) {
	start := strings.Index(field, "((\"")
	if start < 0 {
		return 0, "", false
	}
	rest := field[start+3:]
	end := strings.Index(rest, "\"")
	if end < 0 {
		return 0, "", false
	}
	process := rest[:end]
	pidAt := strings.Index(rest, "pid=")
	if pidAt < 0 {
		return 0, process, true
	}
	pidFields := strings.FieldsFunc(rest[pidAt+4:], func(r rune) bool { return r == ',' || r == ')' || r == ' ' })
	if len(pidFields) == 0 {
		return 0, process, true
	}
	pid, err := strconv.Atoi(pidFields[0])
	if err != nil {
		return 0, process, false
	}
	return pid, process, true
}

// splitHostPort handles the shapes ss/lsof/netstat print: 0.0.0.0:80,
// [::]:80, *:80 and 127.0.0.1:80.
func splitHostPort(value string) (string, int) {
	value = strings.TrimSpace(value)
	colon := strings.LastIndex(value, ":")
	if colon < 0 {
		return value, 0
	}
	host := strings.Trim(value[:colon], "[]")
	port, err := strconv.Atoi(value[colon+1:])
	if err != nil {
		return host, 0
	}
	return host, port
}

// parseLsof decodes `lsof -nP -iTCP -sTCP:LISTEN -F pcn` field output,
// used on macOS. Each process starts with a p<pid> line; c carries the
// command and n the address.
func parseLsof(data string) []Listener {
	var listeners []Listener
	pid, process := 0, ""
	for _, line := range strings.Split(data, "\n") {
		if line == "" {
			continue
		}
		switch line[0] {
		case 'p':
			pid, _ = strconv.Atoi(line[1:])
			process = ""
		case 'c':
			process = line[1:]
		case 'n':
			address, port := splitHostPort(line[1:])
			if port == 0 {
				continue
			}
			listeners = append(listeners, Listener{Port: port, Address: address, PID: pid, Process: process})
		}
	}
	return listeners
}

// parseNetstat decodes Windows `netstat -ano -p tcp` output. The process
// name and working directory are not available from netstat alone.
func parseNetstat(data string) []Listener {
	var listeners []Listener
	for _, line := range strings.Split(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 || !strings.EqualFold(fields[0], "TCP") {
			continue
		}
		if !strings.EqualFold(fields[3], "LISTENING") {
			continue
		}
		address, port := splitHostPort(fields[1])
		if port == 0 {
			continue
		}
		pid, _ := strconv.Atoi(fields[4])
		listeners = append(listeners, Listener{Port: port, Address: address, PID: pid})
	}
	return listeners
}
