// Image support for cloud turns.
//
// A cloud agent sends images the canonical ACP way: a session/update frame
// carries an image content block (base64 data plus a MIME type, sometimes a
// URI). The agent's prose usually points at the picture with ordinary
// Markdown image syntax, e.g.
//
//	Here's the picture of my work:
//
//	![My work](/artifacts/image.png)
//
// Star keeps every image it receives on disk, in a per-session directory, and
// pairs it with the reference the agent wrote. The rendered transcript then
// shows a link to the local file, which a terminal turns into an OSC 8
// hyperlink, so clicking it opens the image in the user's browser.
package cloud

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/charmbracelet/crush/internal/config"
	acp "github.com/coder/acp-go-sdk"
)

// markdownImageRe matches Markdown image references: ![alt](ref).
var markdownImageRe = regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)\)`)

// cloudImage is one image the cloud agent sent, after it was saved locally.
type cloudImage struct {
	// Name is a display name, derived from the ACP URI when there is one.
	Name string
	// Ref is the reference the agent used for it, when known.
	Ref string
	// Path is the local file the bytes were written to.
	Path string
	// URL is the file:// URL for Path.
	URL string
}

// imageStore saves the images of one cloud turn and pairs them with the
// Markdown references the agent writes. It is safe for concurrent use: the
// relay reader is the only writer today, but finishing a turn runs on its own
// goroutine.
type imageStore struct {
	mu      sync.Mutex
	dirPath string
	images  []cloudImage
	byRef   map[string]string
	bound   map[string]bool
}

// newImageStore creates a store rooted at a per-session directory under the
// global data dir, so the files survive the turn and the transcript link keeps
// working when the user clicks it later.
func newImageStore(sessionID string) *imageStore {
	// GlobalConfigData is the data dir's config file; images live beside it.
	dataDir := filepath.Dir(config.GlobalConfigData())
	dir := filepath.Join(dataDir, "cloud-images", sanitizeName(sessionID))
	return &imageStore{
		dirPath: dir,
		byRef:   map[string]string{},
		bound:   map[string]bool{},
	}
}

// save writes an ACP image block to disk and records it in arrival order.
// References the block itself names are bound right away; anything else is
// paired with a reference when the agent's text shows up.
func (s *imageStore) save(img *acp.ContentBlockImage) (cloudImage, error) {
	if img == nil {
		return cloudImage{}, fmt.Errorf("empty image block")
	}
	data, err := decodeImageData(img.Data)
	if err != nil {
		return cloudImage{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dirPath, 0o755); err != nil {
		return cloudImage{}, fmt.Errorf("failed to create image dir: %w", err)
	}
	target := uniquePath(s.dirPath, imageFileName(img, s.images))
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return cloudImage{}, fmt.Errorf("failed to write image: %w", err)
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		abs = target
	}
	saved := cloudImage{
		Name: filepath.Base(abs),
		Path: abs,
		URL:  (&url.URL{Scheme: "file", Path: abs}).String(),
	}
	if img.Uri != nil {
		saved.Ref = *img.Uri
		s.bindLocked(*img.Uri, saved.URL)
		s.bindLocked(filepath.Base(*img.Uri), saved.URL)
	}
	s.images = append(s.images, saved)
	slog.Debug("Saved cloud image", "path", abs, "mime", img.MimeType, "uri", saved.Ref)
	return saved, nil
}

// unclaimed returns the images no reference has been paired with yet, in
// arrival order.
func (s *imageStore) unclaimed() []cloudImage {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []cloudImage
	for _, img := range s.images {
		if !s.bound[img.URL] {
			out = append(out, img)
		}
	}
	return out
}

// claim marks images as shown, so they are not paired with a reference a
// second time.
func (s *imageStore) claim(images []cloudImage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, img := range images {
		s.bound[img.URL] = true
	}
}

// rewrite turns every Markdown image reference that belongs to a received
// image into a clickable link to the local file. References with no image to
// pair with are left untouched. It is idempotent: a replaced reference is a
// plain link, which the image pattern no longer matches.
func (s *imageStore) rewrite(text string) (string, bool) {
	if text == "" || !strings.Contains(text, "![") {
		return text, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	changed := false
	out := markdownImageRe.ReplaceAllStringFunc(text, func(match string) string {
		groups := markdownImageRe.FindStringSubmatch(match)
		alt, ref := groups[1], groups[2]
		img, ok := s.claimLocked(ref)
		if !ok {
			return match
		}
		changed = true
		label := strings.TrimSpace(alt)
		if label == "" {
			label = img.Name
		}
		return "[" + label + "](" + img.URL + ")"
	})
	return out, changed
}

// claimLocked returns the image a reference points at, binding it on first
// use. A reference the block did not name (no URI, or a different path than
// the agent wrote) takes the oldest image still waiting to be shown, which is
// the ACP order the frames arrived in.
func (s *imageStore) claimLocked(ref string) (cloudImage, bool) {
	if target, ok := s.byRef[normRef(ref)]; ok {
		for _, img := range s.images {
			if img.URL == target {
				s.bound[img.URL] = true
				return img, true
			}
		}
	}
	for _, img := range s.images {
		if s.bound[img.URL] {
			continue
		}
		s.bound[img.URL] = true
		s.bindLocked(ref, img.URL)
		img.Ref = ref
		return img, true
	}
	return cloudImage{}, false
}

// bindLocked makes a reference resolve to an image URL.
func (s *imageStore) bindLocked(ref, imageURL string) {
	if ref == "" {
		return
	}
	s.byRef[normRef(ref)] = imageURL
	if base := path.Base(strings.TrimRight(normRef(ref), "/")); base != "" && base != "." && base != "/" {
		if _, exists := s.byRef[base]; !exists {
			s.byRef[base] = imageURL
		}
	}
}

// normRef reduces a Markdown reference to a comparable path: file:// URIs and
// bare paths both collapse to their path, with query strings dropped.
func normRef(ref string) string {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimPrefix(ref, "<")
	ref = strings.TrimSuffix(ref, ">")
	if strings.HasPrefix(ref, "file://") {
		if parsed, err := url.Parse(ref); err == nil {
			ref = parsed.Path
		}
	}
	if idx := strings.IndexByte(ref, '?'); idx >= 0 {
		ref = ref[:idx]
	}
	ref = strings.TrimPrefix(ref, "./")
	return ref
}

// decodeImageData decodes an ACP image payload. Canonical ACP sends bare
// base64; a data URL is accepted too because bridges built on web tooling
// tend to produce one.
func decodeImageData(data string) ([]byte, error) {
	data = strings.TrimSpace(data)
	if strings.HasPrefix(data, "data:") {
		if idx := strings.Index(data, ","); idx >= 0 {
			data = data[idx+1:]
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image data: %w", err)
		}
	}
	if len(decoded) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}
	return decoded, nil
}

// imageFileName picks a file name for an image block: the URI's base name
// when there is one, otherwise a numbered default, always with an extension
// that matches the MIME type.
func imageFileName(img *acp.ContentBlockImage, existing []cloudImage) string {
	ext := extensionForMime(img.MimeType)
	var name string
	if img.Uri != nil {
		name = sanitizeName(path.Base(normRef(*img.Uri)))
	}
	if name == "" || name == "." || name == "/" {
		name = fmt.Sprintf("image-%d", len(existing)+1)
	}
	if filepath.Ext(name) == "" && ext != "" {
		name += ext
	}
	return name
}

// extensionForMime maps the image MIME types ACP can carry to extensions.
func extensionForMime(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/png":
		return ".png"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

// sanitizeName keeps a file name usable inside the session directory.
func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "image"
	}
	var sb strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			sb.WriteRune(r)
		case r == '.' || r == '-' || r == '_':
			sb.WriteRune(r)
		default:
			sb.WriteRune('-')
		}
	}
	cleaned := strings.Trim(sb.String(), "-.")
	if cleaned == "" {
		return "image"
	}
	return cleaned
}

// uniquePath returns dir/name, suffixed when that file already exists so a
// second image with the same name cannot overwrite the first.
func uniquePath(dir, name string) string {
	target := filepath.Join(dir, name)
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return target
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	for i := 2; ; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
