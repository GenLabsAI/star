// Package changes summarizes the modifications in a git working tree so
// an agent or a reviewer can see exactly what a task touched before any
// of it leaves the machine.
package changes

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// File is one modified path in the working tree.
type File struct {
	// Path is the file's current path, relative to the repository root.
	Path string `json:"path"`
	// Status is the two-letter porcelain status, e.g. " M" or "??".
	Status string `json:"status"`
	// Original is the pre-rename path for renames and copies.
	Original string `json:"original,omitempty"`
}

// Summary describes the working tree of a repository.
type Summary struct {
	// Repository reports whether dir is inside a git repository.
	Repository bool `json:"repository"`
	// Root is the repository root, empty outside a repository.
	Root string `json:"root,omitempty"`
	// Branch is the checked out branch, empty on a detached HEAD.
	Branch string `json:"branch,omitempty"`
	// Upstream is the branch's tracking ref, empty when unset.
	Upstream string `json:"upstream,omitempty"`
	// Ahead and Behind count commits relative to Upstream.
	Ahead  int `json:"ahead"`
	Behind int `json:"behind"`
	// Files lists every changed path, including untracked ones.
	Files []File `json:"files"`
	// Stat is git's human-readable diffstat for tracked changes.
	Stat string `json:"stat,omitempty"`
	// PatchBytes is the size of the tracked diff, so callers can decide
	// whether to carry it.
	PatchBytes int `json:"patch_bytes"`
}

// Clean reports whether the working tree has no changes at all.
func (s Summary) Clean() bool {
	return len(s.Files) == 0
}

// Options narrows what a summary reports.
type Options struct {
	// Ignore lists paths that should not count as changes, such as the
	// evidence bundle a verification run just wrote.
	Ignore []string
}

// Collect summarizes the working tree of dir.
func Collect(dir string) (Summary, error) {
	return CollectWith(dir, Options{})
}

// CollectWith summarizes the working tree of dir, leaving out ignored
// paths.
func CollectWith(dir string, opts Options) (Summary, error) {
	root, err := git(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		// Not a repository is a normal answer, not a failure.
		return Summary{}, nil
	}
	summary := Summary{Repository: true, Root: strings.TrimSpace(root)}

	// Branch and upstream are best effort: a detached HEAD or an
	// untracked branch simply reports less.
	if branch, err := git(dir, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		if name := strings.TrimSpace(branch); name != "HEAD" {
			summary.Branch = name
		}
	}
	if upstream, err := git(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}"); err == nil {
		summary.Upstream = strings.TrimSpace(upstream)
		if counts, err := git(dir, "rev-list", "--left-right", "--count", "HEAD...@{upstream}"); err == nil {
			fields := strings.Fields(counts)
			if len(fields) == 2 {
				summary.Ahead, _ = strconv.Atoi(fields[0])
				summary.Behind, _ = strconv.Atoi(fields[1])
			}
		}
	}

	status, err := git(dir, "status", "--porcelain=v1", "-z")
	if err != nil {
		return Summary{}, fmt.Errorf("Failed to read git status: %w", err)
	}
	summary.Files = parsePorcelain(status, summary.Root, opts.Ignore)

	if stat, err := git(dir, "diff", "HEAD", "--stat"); err == nil {
		summary.Stat = strings.TrimRight(stat, "\n")
	}
	if patch, err := Diff(dir); err == nil {
		summary.PatchBytes = len(patch)
	}
	return summary, nil
}

// Diff returns the tracked working-tree diff against HEAD, including
// staged changes. Untracked files are not part of a git diff; callers
// should read them from the summary's file list.
func Diff(dir string) ([]byte, error) {
	out, err := gitBytes(dir, "diff", "HEAD", "--binary")
	if err != nil {
		// An unborn HEAD has nothing to diff against, so fall back to the
		// index.
		out, err = gitBytes(dir, "diff", "--binary")
		if err != nil {
			return nil, fmt.Errorf("Failed to read git diff: %w", err)
		}
	}
	return out, nil
}

// parsePorcelain decodes `git status --porcelain=v1 -z` output. Entries
// are NUL separated; renames and copies carry the original path in the
// next entry. Paths under an ignored directory are dropped.
func parsePorcelain(data, root string, ignore []string) []File {
	entries := strings.Split(data, "\x00")
	files := make([]File, 0, len(entries))
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if len(entry) < 4 {
			continue
		}
		status := entry[:2]
		path := entry[3:]
		file := File{Path: path, Status: status}
		if strings.ContainsAny(status, "RC") && i+1 < len(entries) {
			file.Original = entries[i+1]
			i++
		}
		if ignored(root, path, ignore) {
			continue
		}
		files = append(files, file)
	}
	return files
}

// ignored reports whether a repository-relative path lives under one of
// the ignored directories.
func ignored(root, path string, ignore []string) bool {
	if len(ignore) == 0 || root == "" {
		return false
	}
	for _, target := range ignore {
		rel, err := filepath.Rel(root, target)
		if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
			continue
		}
		rel = filepath.ToSlash(rel)
		if path == rel || strings.HasPrefix(path, rel+"/") {
			return true
		}
	}
	return false
}

func git(dir string, args ...string) (string, error) {
	out, err := gitBytes(dir, args...)
	return string(out), err
}

func gitBytes(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}
