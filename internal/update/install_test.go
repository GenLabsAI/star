package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInstallArchiveStagesAndBoundsVersions(t *testing.T) {
	install := t.TempDir()

	changed, err := installArchive(install, "v0.1.99", testArchive(t, []byte("one")))
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "v0.1.99", readCurrent(install))
	require.FileExists(t, filepath.Join(install, "versions", "v0.1.99", coreBinaryName()))

	time.Sleep(2 * time.Millisecond)
	changed, err = installArchive(install, "v0.1.100", testArchive(t, []byte("two")))
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "v0.1.100", readCurrent(install))

	time.Sleep(2 * time.Millisecond)
	changed, err = installArchive(install, "v0.1.101", testArchive(t, []byte("three")))
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "v0.1.101", readCurrent(install))
	require.FileExists(t, filepath.Join(install, "versions", "v0.1.100", coreBinaryName()))
	require.FileExists(t, filepath.Join(install, "versions", "v0.1.101", coreBinaryName()))
	require.NoDirExists(t, filepath.Join(install, "versions", "v0.1.99"))

	body, err := os.ReadFile(filepath.Join(install, "versions", "v0.1.101", coreBinaryName()))
	require.NoError(t, err)
	require.Equal(t, []byte("three"), body)
}

func TestInstallArchiveIsIdempotent(t *testing.T) {
	install := t.TempDir()

	changed, err := installArchive(install, "v0.1.99", testArchive(t, []byte("first")))
	require.NoError(t, err)
	require.True(t, changed)

	changed, err = installArchive(install, "v0.1.99", testArchive(t, []byte("second")))
	require.NoError(t, err)
	require.False(t, changed)

	body, err := os.ReadFile(filepath.Join(install, "versions", "v0.1.99", coreBinaryName()))
	require.NoError(t, err)
	require.Equal(t, []byte("first"), body)
}

func TestInstallArchiveReadsZip(t *testing.T) {
	install := t.TempDir()

	changed, err := installArchive(install, "v0.1.99", testZip(t, []byte("zipped")))
	require.NoError(t, err)
	require.True(t, changed)

	body, err := os.ReadFile(filepath.Join(install, "versions", "v0.1.99", coreBinaryName()))
	require.NoError(t, err)
	require.Equal(t, []byte("zipped"), body)
}

func TestInstallSkipsDevelopmentBuilds(t *testing.T) {
	t.Setenv("STAR_INSTALL_DIR", t.TempDir())

	_, changed, err := Install(t.Context(), "devel", testClient{"v9.9.9"})
	require.NoError(t, err)
	require.False(t, changed)
}

func TestInstallDownloadsThePlatformAsset(t *testing.T) {
	assetName, ok := assetNameFor(runtime.GOOS, runtime.GOARCH)
	if !ok {
		t.Skipf("no update asset for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	archive := testArchive(t, []byte("downloaded"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	}))
	t.Cleanup(server.Close)

	install := t.TempDir()
	t.Setenv("STAR_INSTALL_DIR", install)
	client := staticClient{release: &Release{
		TagName: "v0.1.100",
		HTMLURL: "https://example.org/release",
		Assets: []Asset{{
			Name:               assetName,
			BrowserDownloadURL: server.URL,
		}},
	}}

	info, changed, err := Install(t.Context(), "v0.1.99", client)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "0.1.100", info.Latest)
	require.Equal(t, "v0.1.100", readCurrent(install))

	body, err := os.ReadFile(filepath.Join(install, "versions", "v0.1.100", coreBinaryName()))
	require.NoError(t, err)
	require.Equal(t, []byte("downloaded"), body)
}

func TestInstallDirUsesEnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STAR_INSTALL_DIR", dir)

	got, ok := installDir()
	require.True(t, ok)
	require.Equal(t, dir, got)
}

func TestAssetNameForPlatform(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
		ok     bool
	}{
		{"windows", "amd64", "star-windows-amd64.zip", true},
		{"linux", "amd64", "star-linux-amd64.tar.gz", true},
		{"linux", "arm64", "star-linux-arm64.tar.gz", true},
		{"darwin", "amd64", "star-darwin-amd64.tar.gz", true},
		{"darwin", "arm64", "star-darwin-arm64.tar.gz", true},
		{"plan9", "amd64", "", false},
	}
	for _, tt := range tests {
		got, ok := assetNameFor(tt.goos, tt.goarch)
		require.Equal(t, tt.ok, ok)
		require.Equal(t, tt.want, got)
	}
}

type staticClient struct {
	release *Release
}

func (c staticClient) Latest(context.Context) (*Release, error) {
	return c.release, nil
}

func testArchive(t *testing.T, core []byte) []byte {
	t.Helper()
	if runtime.GOOS == "windows" {
		return testZip(t, core)
	}
	return testTar(t, core)
}

func testZip(t *testing.T, core []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	header := &zip.FileHeader{Name: "star-" + runtime.GOOS + "-" + runtime.GOARCH + "/" + coreBinaryName()}
	header.SetMode(0o755)
	file, err := writer.CreateHeader(header)
	require.NoError(t, err)
	_, err = file.Write(core)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return buf.Bytes()
}

func testTar(t *testing.T, core []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	writer := tar.NewWriter(gz)
	require.NoError(t, writer.WriteHeader(&tar.Header{
		Name:     "star-" + runtime.GOOS + "-" + runtime.GOARCH + "/" + coreBinaryName(),
		Mode:     0o755,
		Size:     int64(len(core)),
		Typeflag: tar.TypeReg,
	}))
	_, err := writer.Write(core)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}
