package pairing

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseScopes(t *testing.T) {
	t.Parallel()

	set, err := ParseScopes("")
	require.NoError(t, err)
	require.True(t, set.All())
	require.Equal(t, "read,write,exec", set.String())

	set, err = ParseScopes("read, exec")
	require.NoError(t, err)
	require.Equal(t, "read,exec", set.String())
	require.True(t, set.Grants(ScopeExec))
	require.False(t, set.Grants(ScopeWrite))

	set, err = ParseScopes("READ,read")
	require.NoError(t, err)
	require.Equal(t, "read", set.String())

	_, err = ParseScopes("read,root")
	require.ErrorContains(t, err, `Unknown scope "root"`)
}

func TestMintCarriesScopes(t *testing.T) {
	t.Parallel()

	full, err := Mint(AllScopes())
	require.NoError(t, err)
	require.False(t, strings.Contains(full, "."), "an unrestricted token is a bare secret")
	require.True(t, ScopesOf(full).All())

	scoped, err := Mint(ScopeSet{ScopeRead, ScopeExec})
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(scoped, "read.exec."))
	require.Len(t, strings.Split(scoped, "."), 3)
	require.Equal(t, "read,exec", ScopesOf(scoped).String())
	require.True(t, ScopesOf(scoped).Grants(ScopeExec))
	require.False(t, ScopesOf(scoped).Grants(ScopeWrite))

	// A scope list that cannot be read grants nothing rather than
	// everything.
	require.Empty(t, ScopesOf("root.secret"))
}
