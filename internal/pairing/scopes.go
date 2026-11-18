// Scoped pairing: a token can be minted for a subset of what a server
// exposes, so handing a peer read access does not also hand it the shell.
//
// The scopes travel inside the token, in front of the secret and separated
// by dots, e.g. "read.exec.<secret>". A token without a scope prefix is a
// full token, which keeps tokens minted before scopes existed working.
package pairing

import (
	"fmt"
	"strings"
)

// Scope is one capability a token can carry.
type Scope string

// The capabilities a token can be limited to.
const (
	// ScopeRead allows reading what a server exposes: workspaces,
	// sessions, config, health.
	ScopeRead Scope = "read"
	// ScopeWrite allows changing things: creating workspaces and
	// sessions, updating config, answering prompts.
	ScopeWrite Scope = "write"
	// ScopeExec allows running commands and handing the terminal to a
	// process. It is the most powerful scope, so it is the one a server
	// can additionally put behind an approval code.
	ScopeExec Scope = "exec"
)

// tokenScopeSep separates the scope list from the token secret.
const tokenScopeSep = "."

// ScopeSet is the collection of scopes a token carries.
type ScopeSet []Scope

// AllScopes returns every scope, in a stable order.
func AllScopes() ScopeSet {
	return ScopeSet{ScopeRead, ScopeWrite, ScopeExec}
}

// String renders the set the way it is minted and shown: "read,exec".
func (set ScopeSet) String() string {
	parts := make([]string, 0, len(set))
	for _, scope := range set {
		parts = append(parts, string(scope))
	}
	return strings.Join(parts, ",")
}

// Grants reports whether the set contains scope.
func (set ScopeSet) Grants(scope Scope) bool {
	for _, candidate := range set {
		if candidate == scope {
			return true
		}
	}
	return false
}

// All reports whether the set is every scope, i.e. an unrestricted token.
func (set ScopeSet) All() bool {
	return set.Grants(ScopeRead) && set.Grants(ScopeWrite) && set.Grants(ScopeExec)
}

// ParseScopes reads a comma-separated scope list. An empty list means every
// scope, and an unknown name is an error rather than silently ignored: a
// typo must not quietly mint more access than intended.
func ParseScopes(raw string) (ScopeSet, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return AllScopes(), nil
	}
	set := make(ScopeSet, 0, 3)
	for _, field := range strings.Split(raw, ",") {
		scope, err := parseScope(field)
		if err != nil {
			return nil, err
		}
		if !set.Grants(scope) {
			set = append(set, scope)
		}
	}
	return set, nil
}

func parseScope(raw string) (Scope, error) {
	switch Scope(strings.ToLower(strings.TrimSpace(raw))) {
	case ScopeRead:
		return ScopeRead, nil
	case ScopeWrite:
		return ScopeWrite, nil
	case ScopeExec:
		return ScopeExec, nil
	}
	return "", fmt.Errorf("Unknown scope %q: want read, write or exec", strings.TrimSpace(raw))
}

// Mint returns a token carrying scopes. A token for every scope is a bare
// secret, so it looks and behaves exactly like a token from before scopes
// existed.
func Mint(scopes ScopeSet) (string, error) {
	secret, err := NewToken()
	if err != nil {
		return "", err
	}
	if len(scopes) == 0 || scopes.All() {
		return secret, nil
	}
	parts := make([]string, 0, len(scopes)+1)
	for _, scope := range scopes {
		parts = append(parts, string(scope))
	}
	parts = append(parts, secret)
	return strings.Join(parts, tokenScopeSep), nil
}

// ScopesOf reports what a token is allowed to do. A token without a scope
// prefix carries every scope, which is what an unrestricted token has
// always meant.
func ScopesOf(token string) ScopeSet {
	parts := strings.Split(token, tokenScopeSep)
	if len(parts) < 2 {
		return AllScopes()
	}
	set := make(ScopeSet, 0, len(parts)-1)
	for _, part := range parts[:len(parts)-1] {
		scope, err := parseScope(part)
		if err != nil {
			// A token whose prefix cannot be read is treated as
			// unrestricted only when it carries no prefix at all; a
			// garbled prefix grants nothing.
			return ScopeSet{}
		}
		if !set.Grants(scope) {
			set = append(set, scope)
		}
	}
	return set
}
