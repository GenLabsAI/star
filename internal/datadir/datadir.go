// Package datadir resolves where Star keeps a project's data: the
// directory that holds verification bundles, knowledge entries, logs and
// other per-project state.
package datadir

import (
	"path/filepath"

	"github.com/charmbracelet/crush/internal/projects"
)

// DefaultName is the project data directory Star uses when the project
// has no recorded data directory.
const DefaultName = ".crush"

// Resolve returns the data directory for projectDir. A project tracked in
// projects.json uses the directory recorded there, so a custom
// data_directory setting keeps working; everything else falls back to
// <projectDir>/.crush.
func Resolve(projectDir string) string {
	abs, err := filepath.Abs(projectDir)
	if err != nil {
		return filepath.Join(projectDir, DefaultName)
	}
	if list, err := projects.List(); err == nil {
		for _, project := range list {
			if project.Path == abs && project.DataDir != "" {
				return project.DataDir
			}
		}
	}
	return filepath.Join(abs, DefaultName)
}
