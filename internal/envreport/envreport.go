// Package envreport builds the agent-facing environment manifest: the
// predictable filesystem locations, detected project stacks with their
// documented commands, provisioning hooks and runtime readiness checks
// an agent needs before it starts work on a machine.
//
// The manifest never contains credential values. Cloud fields report
// presence only, so the report is safe to paste into a transcript or a
// bug report.
package envreport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/gitutil"
	"github.com/charmbracelet/crush/internal/version"
)

// Manifest is everything an agent can learn about the machine and the
// project without guessing.
type Manifest struct {
	Workspace Workspace `json:"workspace"`
	Star      Star      `json:"star"`
	Git       Git       `json:"git"`
	Project   Project   `json:"project"`
	Hooks     Hooks     `json:"hooks"`
	Runtimes  []Runtime `json:"runtimes"`
	Cloud     Cloud     `json:"cloud"`
	Logs      Logs      `json:"logs"`
}

// Workspace lists the directories Star uses, so an agent never has to
// infer where state lives.
type Workspace struct {
	Dir       string `json:"dir"`
	ConfigDir string `json:"config_dir"`
	StateDir  string `json:"state_dir"`
}

// Star describes the running binary.
type Star struct {
	Version string `json:"version"`
	Commit  string `json:"commit,omitempty"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

// Git describes the repository the workspace belongs to, if any.
type Git struct {
	Repository bool   `json:"repository"`
	Root       string `json:"root,omitempty"`
	Branch     string `json:"branch,omitempty"`
}

// Project is what was detected in the workspace.
type Project struct {
	Stacks       []Stack  `json:"stacks"`
	ContextFiles []string `json:"context_files"`
	Skills       []string `json:"skills"`
	ConfigFiles  []string `json:"config_files"`
}

// Stack is one detected project ecosystem and the commands that set it
// up, build it and test it.
type Stack struct {
	Name      string    `json:"name"`
	Manifests []string  `json:"manifests"`
	Runtimes  []string  `json:"runtimes"`
	Commands  []Command `json:"commands"`
}

// Command is one documented command for a stack.
type Command struct {
	Kind    string `json:"kind"`
	Command string `json:"command"`
}

// Hooks reports the orb lifecycle scripts the project ships. Setup runs
// once on a fresh environment, Resume on every wake-up.
type Hooks struct {
	Setup           string `json:"setup,omitempty"`
	SetupHash       string `json:"setup_hash,omitempty"`
	SetupExecutable bool   `json:"setup_executable"`
	Resume          string `json:"resume,omitempty"`
}

// Runtime is one executable the project needs, with readiness.
type Runtime struct {
	Name       string   `json:"name"`
	Found      bool     `json:"found"`
	Path       string   `json:"path,omitempty"`
	RequiredBy []string `json:"required_by,omitempty"`
}

// Cloud reports which cloud credentials are present. Values are never
// included.
type Cloud struct {
	APIBaseURLSet  bool `json:"api_base_url_set"`
	UserCredential bool `json:"user_credential"`
	WorkerToken    bool `json:"worker_token"`
	ModelReady     bool `json:"model_ready"`
}

// Logs points at the project's log directory when it exists.
type Logs struct {
	Dir   string   `json:"dir,omitempty"`
	Files []string `json:"files,omitempty"`
}

// contextFiles are the project instruction files Star reads into every
// prompt, in the order it looks for them.
var contextFiles = []string{
	"AGENTS.md",
	"CLAUDE.md",
	"CRUSH.md",
	"GEMINI.md",
}

// baseRuntimes are needed by the tooling Star itself relies on in an
// environment.
var baseRuntimes = []string{"git", "sh"}

// Build assembles the manifest for dir.
func Build(dir string) (Manifest, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Manifest{}, fmt.Errorf("Failed to resolve workspace directory: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return Manifest{}, fmt.Errorf("Failed to inspect workspace directory: %w", err)
	}
	if !info.IsDir() {
		return Manifest{}, fmt.Errorf("Workspace path %q is not a directory", abs)
	}

	stacks := DetectStacks(abs)
	commit := version.Commit
	if commit == "unknown" {
		commit = ""
	}
	root := repoRoot(abs)
	branch := ""
	if root != "" {
		branch = gitutil.CurrentBranch(abs)
	}

	return Manifest{
		Workspace: Workspace{
			Dir:       abs,
			ConfigDir: filepath.Dir(config.GlobalConfig()),
			StateDir:  filepath.Dir(config.GlobalConfigData()),
		},
		Star: Star{
			Version: version.Version,
			Commit:  commit,
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		},
		Git:      Git{Repository: root != "", Root: root, Branch: branch},
		Project:  detectProject(abs, stacks),
		Hooks:    detectHooks(abs),
		Runtimes: readiness(stacks),
		Cloud: Cloud{
			APIBaseURLSet:  os.Getenv("DECA_AGENT_API_BASE_URL") != "",
			UserCredential: os.Getenv("DECA_API_KEY") != "" || os.Getenv("DECA_SESSION_ID") != "",
			WorkerToken:    os.Getenv("DECA_AGENT_WORKER_TOKEN") != "",
			ModelReady:     os.Getenv("DECA_MODEL") != "" && os.Getenv("DECA_MODEL_BASE_URL") != "" && os.Getenv("DECA_MODEL_API_KEY") != "",
		},
		Logs: detectLogs(abs),
	}, nil
}

// Missing returns the names of runtimes the project needs but the
// machine does not have.
func (m Manifest) Missing() []string {
	var missing []string
	for _, rt := range m.Runtimes {
		if !rt.Found {
			missing = append(missing, rt.Name)
		}
	}
	return missing
}

func detectProject(dir string, stacks []Stack) Project {
	p := Project{
		Stacks:       stacks,
		ContextFiles: existing(dir, contextFiles),
		Skills:       skills(dir),
		ConfigFiles:  existing(dir, config.ProjectConfigs(dir)),
	}
	slices.Sort(p.ConfigFiles)
	return p
}

func detectHooks(dir string) Hooks {
	var h Hooks
	setupPath := filepath.Join(dir, ".agents", "setup")
	if data, err := os.ReadFile(setupPath); err == nil && len(data) > 0 {
		sum := sha256.Sum256(data)
		h.Setup = setupPath
		h.SetupHash = hex.EncodeToString(sum[:])
		if info, statErr := os.Stat(setupPath); statErr == nil {
			h.SetupExecutable = info.Mode()&0o111 != 0
		}
	}
	resumePath := filepath.Join(dir, ".agents", "resume")
	if info, err := os.Stat(resumePath); err == nil && info.Size() > 0 {
		h.Resume = resumePath
	}
	return h
}

func detectLogs(dir string) Logs {
	logDir := filepath.Join(dir, ".crush", "logs")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return Logs{}
	}
	logs := Logs{Dir: logDir}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}
		logs.Files = append(logs.Files, filepath.Join(logDir, entry.Name()))
	}
	return logs
}

// skills lists project skill directories, which Star loads as reusable
// instructions.
func skills(dir string) []string {
	entries, err := os.ReadDir(filepath.Join(dir, ".agents", "skills"))
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", entry.Name(), "SKILL.md")); err == nil {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names
}

// readiness resolves every runtime the detected stacks need, plus the
// base tooling, and records which stacks asked for each.
func readiness(stacks []Stack) []Runtime {
	needed := map[string][]string{}
	for _, name := range baseRuntimes {
		needed[name] = nil
	}
	for _, stack := range stacks {
		for _, name := range stack.Runtimes {
			if !slices.Contains(needed[name], stack.Name) {
				needed[name] = append(needed[name], stack.Name)
			}
		}
	}
	runtimes := make([]Runtime, 0, len(needed))
	for name, by := range needed {
		rt := Runtime{Name: name, RequiredBy: by}
		if path, err := exec.LookPath(name); err == nil {
			rt.Found = true
			rt.Path = path
		}
		runtimes = append(runtimes, rt)
	}
	sort.Slice(runtimes, func(i, j int) bool { return runtimes[i].Name < runtimes[j].Name })
	return runtimes
}

// DetectStacks inspects dir and returns the project ecosystems it finds,
// with the commands an agent should use for each. The order is stable so
// reports and tests are comparable.
func DetectStacks(dir string) []Stack {
	var stacks []Stack
	if has(dir, "go.mod") {
		stack := Stack{
			Name:      "Go",
			Manifests: []string{filepath.Join(dir, "go.mod")},
			Runtimes:  []string{"go"},
			Commands: []Command{
				{Kind: "install", Command: "go mod download"},
				{Kind: "build", Command: "go build ./..."},
				{Kind: "test", Command: "go test ./..."},
			},
		}
		if has(dir, "go.sum") {
			stack.Manifests = append(stack.Manifests, filepath.Join(dir, "go.sum"))
		}
		stacks = append(stacks, stack)
	}
	if has(dir, "package.json") {
		stacks = append(stacks, nodeStack(dir))
	}
	if has(dir, "Cargo.toml") {
		stack := Stack{
			Name:      "Rust",
			Manifests: []string{filepath.Join(dir, "Cargo.toml")},
			Runtimes:  []string{"cargo"},
			Commands: []Command{
				{Kind: "install", Command: "cargo fetch"},
				{Kind: "build", Command: "cargo build"},
				{Kind: "test", Command: "cargo test"},
			},
		}
		if has(dir, "Cargo.lock") {
			stack.Manifests = append(stack.Manifests, filepath.Join(dir, "Cargo.lock"))
		}
		stacks = append(stacks, stack)
	}
	if stack, ok := pythonStack(dir); ok {
		stacks = append(stacks, stack)
	}
	if has(dir, "compose.yaml") || has(dir, "compose.yml") || has(dir, "docker-compose.yaml") || has(dir, "docker-compose.yml") {
		stack := Stack{
			Name:      "Docker Compose",
			Runtimes:  []string{"docker"},
			Commands:  []Command{{Kind: "services", Command: "docker compose up -d"}},
			Manifests: existing(dir, []string{"compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"}),
		}
		stacks = append(stacks, stack)
	}
	return stacks
}

func nodeStack(dir string) Stack {
	pm, install := "npm", "npm install"
	switch {
	case has(dir, "bun.lock"), has(dir, "bun.lockb"):
		pm, install = "bun", "bun install"
	case has(dir, "pnpm-lock.yaml"):
		pm, install = "pnpm", "pnpm install"
	case has(dir, "yarn.lock"):
		pm, install = "yarn", "yarn install"
	case has(dir, "package-lock.json"), has(dir, "npm-shrinkwrap.json"):
		pm, install = "npm", "npm ci"
	}
	stack := Stack{
		Name:      "Node",
		Manifests: []string{filepath.Join(dir, "package.json")},
		Runtimes:  []string{"node", pm},
		Commands:  []Command{{Kind: "install", Command: install}},
	}
	scripts := packageScripts(dir)
	if scripts["build"] != "" {
		stack.Commands = append(stack.Commands, Command{Kind: "build", Command: pm + " run build"})
	}
	if scripts["test"] != "" {
		stack.Commands = append(stack.Commands, Command{Kind: "test", Command: pm + " test"})
	}
	if scripts["dev"] != "" {
		stack.Commands = append(stack.Commands, Command{Kind: "dev", Command: pm + " run dev"})
	} else if scripts["start"] != "" {
		stack.Commands = append(stack.Commands, Command{Kind: "dev", Command: pm + " run start"})
	}
	for _, lock := range []string{"bun.lock", "bun.lockb", "pnpm-lock.yaml", "yarn.lock", "package-lock.json", "npm-shrinkwrap.json"} {
		if has(dir, lock) {
			stack.Manifests = append(stack.Manifests, filepath.Join(dir, lock))
		}
	}
	return stack
}

func pythonStack(dir string) (Stack, bool) {
	manifests := existing(dir, []string{"pyproject.toml", "requirements.txt", "setup.py", "uv.lock"})
	if len(manifests) == 0 {
		return Stack{}, false
	}
	install := "python3 -m pip install -e ."
	switch {
	case has(dir, "uv.lock"):
		install = "uv sync"
	case has(dir, "requirements.txt"):
		install = "python3 -m pip install -r requirements.txt"
	}
	stack := Stack{
		Name:      "Python",
		Manifests: manifests,
		Runtimes:  []string{"python3"},
		Commands:  []Command{{Kind: "install", Command: install}},
	}
	if has(dir, "uv.lock") {
		stack.Runtimes = append(stack.Runtimes, "uv")
	}
	return stack, true
}

// packageScripts reads the scripts block of package.json. A missing or
// unreadable file yields no scripts, which simply documents fewer
// commands.
func packageScripts(dir string) map[string]string {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}
	return pkg.Scripts
}

// repoRoot walks up from dir looking for a repository, resolving the
// "gitdir:" pointer that worktrees and submodules use in a .git file.
func repoRoot(dir string) string {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func has(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

// existing returns the subset of names present in dir, as absolute
// paths.
func existing(dir string, names []string) []string {
	var found []string
	for _, name := range names {
		path := name
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, name)
		}
		if has(filepath.Dir(path), filepath.Base(path)) {
			found = append(found, path)
		}
	}
	return found
}
