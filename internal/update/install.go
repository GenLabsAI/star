package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	maxAssetSize    = 512 << 20
	stagingLifetime = time.Hour
)

// Install checks for a newer release and, when one is available, downloads it
// into the versioned install directory. The running process keeps using the
// core it started with; the launcher picks up the new version on the next
// launch.
//
// The returned bool reports whether this call installed or activated a new
// version. A release that was already staged by another window is not reported
// again.
func Install(ctx context.Context, current string, client Client) (Info, bool, error) {
	release, err := client.Latest(ctx)
	if err != nil {
		return Info{Current: current, Latest: current}, false, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	info := infoFromRelease(current, release)
	if current == "" || !info.Available() || info.IsDevelopment() {
		return info, false, nil
	}

	installDir, ok := installDir()
	if !ok {
		return info, false, nil
	}

	assetName, ok := assetNameFor(runtime.GOOS, runtime.GOARCH)
	if !ok {
		return info, false, fmt.Errorf("no update asset for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	var assetURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			assetURL = asset.BrowserDownloadURL
			break
		}
	}
	if assetURL == "" {
		return info, false, fmt.Errorf("release has no %s asset for this platform", assetName)
	}

	version := strings.TrimSpace(release.TagName)
	if !safeVersion(version) {
		return info, false, fmt.Errorf("release version is not safe to use as a directory: %s", version)
	}

	changed, err := installVersion(ctx, installDir, version, assetURL, current)
	return info, changed, err
}

func installVersion(ctx context.Context, installDir, version, assetURL, runningVersion string) (bool, error) {
	dest := filepath.Join(installDir, "versions", version, coreBinaryName())
	if fileExists(dest) {
		return activateInstalled(installDir, version, runningVersion)
	}

	archive, err := downloadAsset(ctx, assetURL)
	if err != nil {
		return false, err
	}
	return installArchive(installDir, version, archive, runningVersion)
}

func activateInstalled(installDir, version string, keep ...string) (bool, error) {
	alreadyActive := readCurrent(installDir) == version
	if err := writeCurrent(installDir, version); err != nil {
		return false, err
	}
	cleanupInstall(installDir, keep...)
	return !alreadyActive, nil
}

func installArchive(installDir, version string, archive []byte, keep ...string) (bool, error) {
	if !safeVersion(version) {
		return false, fmt.Errorf("release version is not safe to use as a directory: %s", version)
	}

	versionsDir := filepath.Join(installDir, "versions")
	destDir := filepath.Join(versionsDir, version)
	dest := filepath.Join(destDir, coreBinaryName())
	if fileExists(dest) {
		return activateInstalled(installDir, version, keep...)
	}

	if err := os.MkdirAll(versionsDir, 0o755); err != nil {
		return false, fmt.Errorf("could not create versions directory: %w", err)
	}
	staging, err := os.MkdirTemp(versionsDir, ".staging-"+version+"-")
	if err != nil {
		return false, fmt.Errorf("could not create staging directory: %w", err)
	}
	keepStaging := false
	defer func() {
		if !keepStaging {
			_ = os.RemoveAll(staging)
		}
	}()

	if err := extractCore(archive, filepath.Join(staging, coreBinaryName())); err != nil {
		return false, err
	}

	// A valid version may have been installed by another Star window while this
	// one was downloading. Never remove it out from under that window.
	if fileExists(dest) {
		return activateInstalled(installDir, version, keep...)
	}
	if _, err := os.Stat(destDir); err == nil {
		if fileExists(dest) {
			return activateInstalled(installDir, version, keep...)
		}
		// A partial directory from an interrupted attempt has no usable core.
		_ = os.RemoveAll(destDir)
	}

	var renameErr error
	for attempt := 0; attempt < 10; attempt++ {
		if err := os.Rename(staging, destDir); err == nil {
			keepStaging = true
			break
		} else {
			renameErr = err
		}
		if fileExists(dest) {
			// Another process won the race.
			return activateInstalled(installDir, version, keep...)
		}
		time.Sleep(time.Duration(100*(attempt+1)) * time.Millisecond)
	}
	if !keepStaging {
		return false, fmt.Errorf("could not activate version %s: %w", version, renameErr)
	}

	if err := writeCurrent(installDir, version); err != nil {
		return false, err
	}
	cleanupInstall(installDir, keep...)
	return true, nil
}

func downloadAsset(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := (&http.Client{Timeout: 10 * time.Minute}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAssetSize+1))
	if err != nil {
		return nil, fmt.Errorf("could not read download: %w", err)
	}
	if len(body) > maxAssetSize {
		return nil, errors.New("downloaded update is too large")
	}
	if len(body) == 0 {
		return nil, errors.New("download returned no data")
	}
	return body, nil
}

func extractCore(archive []byte, dest string) error {
	if bytes.HasPrefix(archive, []byte("PK\x03\x04")) || bytes.HasPrefix(archive, []byte("PK\x05\x06")) {
		return extractZipCore(archive, dest)
	}
	return extractTarCore(archive, dest)
}

func extractZipCore(archive []byte, dest string) error {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return fmt.Errorf("could not open the downloaded archive: %w", err)
	}
	wanted := coreBinaryName()
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		name := strings.ReplaceAll(file.Name, "\\", "/")
		if path.Base(name) != wanted {
			continue
		}
		body, err := file.Open()
		if err != nil {
			return fmt.Errorf("could not read %s from the archive: %w", wanted, err)
		}
		err = writeCore(dest, body)
		_ = body.Close()
		return err
	}
	return fmt.Errorf("%s was not in the downloaded archive", wanted)
}

func extractTarCore(archive []byte, dest string) error {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return fmt.Errorf("could not open the downloaded archive: %w", err)
	}
	defer gz.Close()

	reader := tar.NewReader(gz)
	wanted := coreBinaryName()
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("could not read the downloaded archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}
		if path.Base(strings.ReplaceAll(header.Name, "\\", "/")) != wanted {
			continue
		}
		return writeCore(dest, io.LimitReader(reader, header.Size))
	}
	return fmt.Errorf("%s was not in the downloaded archive", wanted)
}

func writeCore(dest string, body io.Reader) error {
	file, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("could not create the new core: %w", err)
	}
	_, copyErr := io.Copy(file, body)
	syncErr := file.Sync()
	closeErr := file.Close()
	if copyErr != nil {
		return fmt.Errorf("could not write the new core: %w", copyErr)
	}
	if syncErr != nil {
		return fmt.Errorf("could not flush the new core: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("could not close the new core: %w", closeErr)
	}
	_ = os.Chmod(dest, 0o755)
	return nil
}

func installDir() (string, bool) {
	if dir := strings.TrimSpace(os.Getenv("STAR_INSTALL_DIR")); dir != "" {
		return dir, true
	}

	exe, err := os.Executable()
	if err != nil {
		return "", false
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	dir := filepath.Dir(exe)
	if filepath.Base(filepath.Dir(dir)) == "versions" {
		return filepath.Dir(filepath.Dir(dir)), true
	}
	if filepath.Base(dir) == "bin" {
		return dir, true
	}
	return "", false
}

func coreBinaryName() string {
	if runtime.GOOS == "windows" {
		return "star-core.exe"
	}
	return "star-core"
}

func assetNameFor(goos, goarch string) (string, bool) {
	switch {
	case goos == "windows" && goarch == "amd64":
		return "star-windows-amd64.zip", true
	case goos == "linux" && goarch == "amd64":
		return "star-linux-amd64.tar.gz", true
	case goos == "linux" && goarch == "arm64":
		return "star-linux-arm64.tar.gz", true
	case goos == "darwin" && goarch == "amd64":
		return "star-darwin-amd64.tar.gz", true
	case goos == "darwin" && goarch == "arm64":
		return "star-darwin-arm64.tar.gz", true
	default:
		return "", false
	}
}

func readCurrent(installDir string) string {
	raw, err := os.ReadFile(filepath.Join(installDir, "current"))
	if err != nil {
		return ""
	}
	version := strings.TrimSpace(string(raw))
	if !safeVersion(version) {
		return ""
	}
	return version
}

func writeCurrent(installDir, version string) error {
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return fmt.Errorf("could not create install directory: %w", err)
	}
	tmp, err := os.CreateTemp(installDir, ".current-*")
	if err != nil {
		return fmt.Errorf("could not create version pointer: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.WriteString(version); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("could not write version pointer: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("could not flush version pointer: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("could not close version pointer: %w", err)
	}
	if err := os.Rename(tmpName, filepath.Join(installDir, "current")); err == nil {
		return nil
	}
	// Windows and some filesystems refuse to replace an existing file.
	if err := os.Remove(filepath.Join(installDir, "current")); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not replace version pointer: %w", err)
	}
	if err := os.Rename(tmpName, filepath.Join(installDir, "current")); err != nil {
		return fmt.Errorf("could not activate version: %w", err)
	}
	return nil
}

func cleanupInstall(installDir string, keep ...string) {
	versionsDir := filepath.Join(installDir, "versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return
	}

	current := readCurrent(installDir)
	kept := map[string]bool{}
	for _, version := range append([]string{current}, keep...) {
		version = normalizeVersion(version)
		if safeVersion(version) {
			kept[version] = true
		}
	}

	type candidate struct {
		name     string
		path     string
		modified time.Time
	}
	var candidates []candidate
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if strings.HasPrefix(name, ".staging-") {
			if time.Since(info.ModTime()) > stagingLifetime {
				_ = os.RemoveAll(filepath.Join(versionsDir, name))
			}
			continue
		}
		if safeVersion(name) {
			candidates = append(candidates, candidate{name, filepath.Join(versionsDir, name), info.ModTime()})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].modified.After(candidates[j].modified)
	})
	for _, candidate := range candidates {
		if len(kept) >= 2 {
			break
		}
		kept[candidate.name] = true
	}
	for _, candidate := range candidates {
		if !kept[candidate.name] {
			_ = os.RemoveAll(candidate.path)
		}
	}

	_ = os.Remove(filepath.Join(installDir, ".star-core-update.tmp"))
	_ = os.Remove(filepath.Join(installDir, ".star-core-old.bak"))
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" || strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func safeVersion(version string) bool {
	return version != "" &&
		len(version) <= 64 &&
		version != "." &&
		version != ".." &&
		strings.IndexFunc(version, func(r rune) bool {
			return !(r >= 'a' && r <= 'z' ||
				r >= 'A' && r <= 'Z' ||
				r >= '0' && r <= '9' ||
				r == '.' || r == '-' || r == '_')
		}) == -1
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
