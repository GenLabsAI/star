// Package pairing mints and checks the bearer token a remote peer
// presents to a Star server.
//
// Without pairing, a Star server over TCP is protected only from browser
// pages. A pairing token makes the server prove a peer was authorized by
// the user: requests without the token get 401 and never reach the API.
package pairing

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/crush/internal/config"
)

// Environment variables a peer uses to present its token.
const (
	// EnvToken holds a pairing token directly.
	EnvToken = "STAR_SERVER_TOKEN"
	// EnvTokenFile points at a file holding the token.
	EnvTokenFile = "STAR_SERVER_TOKEN_FILE"
)

// FileName is the conventional token file name in the config directory.
const FileName = "pairing-token"

// tokenBytes is the entropy behind one token.
const tokenBytes = 32

// NewToken returns a fresh 256-bit pairing token in URL-safe base64, so it
// can travel through a shell, a query string or a chat message without
// escaping.
func NewToken() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("Failed to generate pairing token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// DefaultPath returns the conventional token file location in the config
// directory.
func DefaultPath() string {
	return filepath.Join(filepath.Dir(config.GlobalConfig()), FileName)
}

// Save writes a token to path with owner-only permissions.
func Save(path, token string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("Failed to create token directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return fmt.Errorf("Failed to write pairing token: %w", err)
	}
	return nil
}

// Read returns the token stored at path.
func Read(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Failed to read pairing token: %w", err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("Pairing token file %s is empty", path)
	}
	return token, nil
}

// Resolve returns the token this process should present, and where it
// came from. Both are empty when the peer has not been paired.
func Resolve() (token, source string) {
	if fromEnv := strings.TrimSpace(os.Getenv(EnvToken)); fromEnv != "" {
		return fromEnv, "env " + EnvToken
	}
	if path := strings.TrimSpace(os.Getenv(EnvTokenFile)); path != "" {
		if token, err := Read(path); err == nil {
			return token, path
		}
	}
	return "", ""
}

// Match compares a presented token with the expected one without leaking
// their contents through timing.
func Match(presented, expected string) bool {
	if presented == "" || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(presented), []byte(expected)) == 1
}
