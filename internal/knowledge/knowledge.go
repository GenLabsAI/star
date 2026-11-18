// Package knowledge stores durable project discoveries that outlive a
// session: quirks, commands, architecture notes and anything an agent
// worked out once and should not have to work out again.
//
// Each entry is one markdown file in the project's data directory, so it
// survives restarts, is readable with any editor, and can be versioned
// with the project if the user wants that. Content is stored verbatim
// and returned verbatim.
package knowledge

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/datadir"
)

// DirName is the knowledge directory inside a project's data directory.
const DirName = "knowledge"

// MaxBody bounds one entry, so a pasted log cannot bloat the store
// without a deliberate decision.
const MaxBody = 256 << 10

// namePattern is the accepted entry name: a slug, never a path.
var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

// Entry is one stored note.
type Entry struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	Updated time.Time `json:"updated"`
	// Summary is the entry's first non-empty line, so a listing is
	// useful without opening every file.
	Summary string `json:"summary"`
}

// Store is the knowledge directory for one project.
type Store struct {
	dir string
}

// NewStore returns the store for a project directory.
func NewStore(projectDir string) *Store {
	return &Store{dir: filepath.Join(datadir.Resolve(projectDir), DirName)}
}

// Dir returns the knowledge directory. It may not exist yet.
func (s *Store) Dir() string {
	return s.dir
}

// Path returns the file path for a name.
func (s *Store) Path(name string) string {
	return filepath.Join(s.dir, name+".md")
}

// Add writes an entry. Bodies are stored verbatim apart from a single
// trailing newline, so a reader gets back what the writer wrote. An
// existing entry is replaced: the newest answer wins.
func (s *Store) Add(name, body string) (Entry, error) {
	if err := ValidateName(name); err != nil {
		return Entry{}, err
	}
	if strings.TrimSpace(body) == "" {
		return Entry{}, errors.New("Entry body is empty")
	}
	if len(body) > MaxBody {
		return Entry{}, fmt.Errorf("Entry is larger than %d bytes", MaxBody)
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return Entry{}, fmt.Errorf("Failed to create knowledge directory: %w", err)
	}

	content := strings.TrimRight(body, "\n") + "\n"
	path := s.Path(name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return Entry{}, fmt.Errorf("Failed to write knowledge entry: %w", err)
	}
	return s.entry(name)
}

// Get returns an entry's body exactly as stored.
func (s *Store) Get(name string) (Entry, string, error) {
	if err := ValidateName(name); err != nil {
		return Entry{}, "", err
	}
	data, err := os.ReadFile(s.Path(name))
	if err != nil {
		if os.IsNotExist(err) {
			return Entry{}, "", fmt.Errorf("No knowledge entry named %q", name)
		}
		return Entry{}, "", fmt.Errorf("Failed to read knowledge entry: %w", err)
	}
	entry, err := s.entry(name)
	if err != nil {
		return Entry{}, "", err
	}
	return entry, string(data), nil
}

// Remove deletes an entry.
func (s *Store) Remove(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	if err := os.Remove(s.Path(name)); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("No knowledge entry named %q", name)
		}
		return fmt.Errorf("Failed to remove knowledge entry: %w", err)
	}
	return nil
}

// List returns every entry, ordered by name. A missing directory is an
// empty store, not an error.
func (s *Store) List() ([]Entry, error) {
	dirEntries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("Failed to read knowledge directory: %w", err)
	}
	var entries []Entry
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() || !strings.HasSuffix(dirEntry.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(dirEntry.Name(), ".md")
		if entry, err := s.entry(name); err == nil {
			entries = append(entries, entry)
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries, nil
}

func (s *Store) entry(name string) (Entry, error) {
	path := s.Path(name)
	info, err := os.Stat(path)
	if err != nil {
		return Entry{}, fmt.Errorf("Failed to inspect knowledge entry: %w", err)
	}
	entry := Entry{
		Name:    name,
		Path:    path,
		Size:    info.Size(),
		Updated: info.ModTime().UTC(),
	}
	if data, err := os.ReadFile(path); err == nil {
		entry.Summary = summary(string(data))
	}
	return entry, nil
}

// summary returns the first line with content, trimmed for display.
func summary(body string) string {
	for _, line := range strings.Split(body, "\n") {
		text := strings.TrimSpace(strings.TrimLeft(line, "#"))
		if text == "" {
			continue
		}
		if len(text) > 120 {
			text = text[:117] + "..."
		}
		return text
	}
	return ""
}

// ValidateName rejects anything that is not a plain slug, so an entry can
// never escape the knowledge directory.
func ValidateName(name string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("Invalid entry name %q: use lowercase letters, digits, '.', '_' or '-' (max 64)", name)
	}
	return nil
}
