// Approval codes: a short-lived code an operator mints on the server host
// to let one sensitive request through. The pairing token proves a peer was
// authorized once; an approval code proves a person is present now.
package pairing

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ApprovalHeader is the header a peer presents an approval code in.
const ApprovalHeader = "X-Star-Approval"

// DefaultApprovalTTL is how long a minted approval code stays valid.
const DefaultApprovalTTL = 5 * time.Minute

// approvalBytes is the entropy behind one approval code.
const approvalBytes = 5

// Approval is a minted code and when it stops being valid.
type Approval struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewApproval returns a fresh approval code. It is short because a person
// copies it by hand, and 40 random bits is far too many to guess within the
// minutes it lives.
func NewApproval() (string, error) {
	buf := make([]byte, approvalBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("Failed to generate approval code: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), nil
}

// SaveApproval mints a code and writes it, with its expiry, to path with
// owner-only permissions. The file is what a running server checks, so an
// approval can be granted without restarting it.
func SaveApproval(path string, ttl time.Duration) (Approval, error) {
	code, err := NewApproval()
	if err != nil {
		return Approval{}, err
	}
	if ttl <= 0 {
		ttl = DefaultApprovalTTL
	}
	approval := Approval{Code: code, ExpiresAt: time.Now().Add(ttl).UTC()}
	data, err := json.MarshalIndent(approval, "", "  ")
	if err != nil {
		return Approval{}, fmt.Errorf("Failed to encode approval: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return Approval{}, fmt.Errorf("Failed to create approval directory: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return Approval{}, fmt.Errorf("Failed to write approval: %w", err)
	}
	return approval, nil
}

// ReadApproval returns the approval stored at path.
func ReadApproval(path string) (Approval, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Approval{}, fmt.Errorf("Failed to read approval: %w", err)
	}
	var approval Approval
	if err := json.Unmarshal(data, &approval); err != nil {
		return Approval{}, fmt.Errorf("Failed to decode approval %s: %w", path, err)
	}
	if strings.TrimSpace(approval.Code) == "" {
		return Approval{}, fmt.Errorf("Approval file %s holds no code", path)
	}
	return approval, nil
}

// Valid reports whether code matches this approval and has not expired.
func (a Approval) Valid(code string, now time.Time) bool {
	if a.Code == "" || code == "" {
		return false
	}
	if !now.Before(a.ExpiresAt) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(code)), []byte(a.Code)) == 1
}
