package cloud

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/require"
)

func pngBlock(t *testing.T, uri string) *acp.ContentBlockImage {
	t.Helper()
	data := []byte("\x89PNG\r\n\x1a\nfake-image-bytes")
	block := &acp.ContentBlockImage{
		Data:     base64.StdEncoding.EncodeToString(data),
		MimeType: "image/png",
		Type:     "image",
	}
	if uri != "" {
		block.Uri = &uri
	}
	return block
}

// The agent writes a Markdown image with the same path it names on the image
// block; the transcript should point at the saved file instead.
func TestImageStoreLinksMarkdownReferenceToSavedFile(t *testing.T) {
	t.Setenv("CRUSH_GLOBAL_DATA", t.TempDir())
	store := newImageStore("session-1")

	saved, err := store.save(pngBlock(t, "/artifacts/image.png"))
	require.NoError(t, err)
	require.FileExists(t, saved.Path)
	require.Equal(t, ".png", filepath.Ext(saved.Path))
	require.True(t, strings.HasPrefix(saved.URL, "file://"))

	text := "Here's the picture of my work:\n\n![My work](/artifacts/image.png)\n"
	rewritten, changed := store.rewrite(text)
	require.True(t, changed)
	require.Contains(t, rewritten, "[My work]("+saved.URL+")")
	require.NotContains(t, rewritten, "![My work]")
}

// A bridge that sends the image without a URI still pairs with the reference
// in ACP order, and a second pass must not rewrite what is already a link.
func TestImageStorePairsByArrivalOrderAndIsIdempotent(t *testing.T) {
	t.Setenv("CRUSH_GLOBAL_DATA", t.TempDir())
	store := newImageStore("session-2")

	first, err := store.save(pngBlock(t, ""))
	require.NoError(t, err)
	_, err = store.save(pngBlock(t, ""))
	require.NoError(t, err)

	text := "![one](/artifacts/one.png) and ![two](/artifacts/two.png)"
	rewritten, changed := store.rewrite(text)
	require.True(t, changed)
	require.Contains(t, rewritten, "[one]("+first.URL+")")
	require.Contains(t, rewritten, "[two](file://")

	again, changedAgain := store.rewrite(rewritten)
	require.False(t, changedAgain)
	require.Equal(t, rewritten, again)
}

// Alt text falls back to the saved file name, and an image buffer that is not
// base64 is rejected instead of writing junk.
func TestImageStoreAltFallbackAndBadData(t *testing.T) {
	t.Setenv("CRUSH_GLOBAL_DATA", t.TempDir())
	store := newImageStore("session-3")

	saved, err := store.save(pngBlock(t, ""))
	require.NoError(t, err)
	rewritten, changed := store.rewrite("![](/artifacts/shot.png)")
	require.True(t, changed)
	require.Contains(t, rewritten, "["+saved.Name+"]("+saved.URL+")")

	_, err = store.save(&acp.ContentBlockImage{Data: "not base64!!", MimeType: "image/png"})
	require.Error(t, err)
}

// Name collisions inside a session directory keep both files.
func TestImageStoreKeepsBothCollidingNames(t *testing.T) {
	t.Setenv("CRUSH_GLOBAL_DATA", t.TempDir())
	store := newImageStore("session-4")

	first, err := store.save(pngBlock(t, "/artifacts/hero.png"))
	require.NoError(t, err)
	second, err := store.save(pngBlock(t, "/artifacts/hero.png"))
	require.NoError(t, err)
	require.NotEqual(t, first.Path, second.Path)
	require.FileExists(t, first.Path)
	require.FileExists(t, second.Path)
}

// The session directory is derived from the Star session id, sanitized so it
// cannot escape the data dir.
func TestImageStoreSanitizesSessionDir(t *testing.T) {
	data := t.TempDir()
	t.Setenv("CRUSH_GLOBAL_DATA", data)
	store := newImageStore("../../etc/passwd")

	saved, err := store.save(pngBlock(t, ""))
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(saved.Path, filepath.Join(data, "cloud-images")))
	_, statErr := os.Stat(saved.Path)
	require.NoError(t, statErr)
}
