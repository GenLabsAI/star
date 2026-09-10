Create an isolated git worktree for parallel work without file conflicts.

Usage:
- `branch`: Optional branch name for the worktree. If omitted, a detached HEAD is used.

The worktree is created under the project's `.crush/worktrees/` directory and linked to the current repository. Use `exit_worktree` to tear it down when done.
