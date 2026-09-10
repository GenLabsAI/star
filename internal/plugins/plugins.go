package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

const ManifestFile = "star-plugin.json"

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

type Manifest struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description,omitempty"`
	Skills      []string          `json:"skills,omitempty"`
	Hooks       map[string][]Hook `json:"hooks,omitempty"`
	MCP         map[string]any    `json:"mcp,omitempty"`
	LSP         map[string]any    `json:"lsp,omitempty"`
	Agents      map[string]any    `json:"agents,omitempty"`
}

type Hook struct {
	Name    string `json:"name,omitempty"`
	Matcher string `json:"matcher,omitempty"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

type Installed struct {
	Manifest
	Path    string `json:"path"`
	Enabled bool   `json:"enabled"`
}

func Load(path string) (Manifest, error) {
	data, err := os.ReadFile(filepath.Join(path, ManifestFile))
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse plugin manifest: %w", err)
	}
	if !validName.MatchString(manifest.Name) {
		return Manifest{}, errors.New("plugin name must contain only letters, numbers, dots, underscores, or hyphens")
	}
	if manifest.Version == "" {
		return Manifest{}, errors.New("plugin version is required")
	}
	return manifest, nil
}

func Install(source, directory string) (Installed, error) {
	info, err := os.Stat(source)
	if err != nil {
		return Installed{}, fmt.Errorf("access plugin source: %w", err)
	}
	if !info.IsDir() {
		return Installed{}, errors.New("plugin source must be a directory")
	}
	manifest, err := Load(source)
	if err != nil {
		return Installed{}, err
	}
	destination := filepath.Join(directory, manifest.Name)
	if _, err := os.Stat(destination); err == nil {
		return Installed{}, fmt.Errorf("plugin %q is already installed", manifest.Name)
	}
	if err := copyDir(source, destination); err != nil {
		return Installed{}, err
	}
	return Installed{Manifest: manifest, Path: destination, Enabled: true}, nil
}

func List(directory string) ([]Installed, error) {
	entries, err := os.ReadDir(directory)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	result := make([]Installed, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		manifest, err := Load(path)
		if err != nil {
			continue
		}
		_, disabledErr := os.Stat(filepath.Join(path, ".disabled"))
		result = append(result, Installed{Manifest: manifest, Path: path, Enabled: errors.Is(disabledErr, fs.ErrNotExist)})
	}
	return result, nil
}

func Remove(name, directory string) error {
	if !validName.MatchString(name) {
		return errors.New("invalid plugin name")
	}
	return os.RemoveAll(filepath.Join(directory, name))
}

func SetEnabled(name, directory string, enabled bool) error {
	path := filepath.Join(directory, name)
	if _, err := Load(path); err != nil {
		return err
	}
	marker := filepath.Join(path, ".disabled")
	if enabled {
		if err := os.Remove(marker); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	return os.WriteFile(marker, nil, 0o644)
}

func copyDir(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
