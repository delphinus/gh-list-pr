package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const cwdFileEnv = "GH_LIST_PR_CWD_FILE"

type worktree struct {
	path   string
	branch string // local branch name; "" when detached or bare
	bare   bool
}

// parseWorktrees parses `git worktree list --porcelain`. The first entry is
// always the main worktree.
func parseWorktrees(out string) []worktree {
	var result []worktree
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			result = append(result, worktree{path: strings.TrimPrefix(line, "worktree ")})
		case len(result) == 0:
		case strings.HasPrefix(line, "branch refs/heads/"):
			result[len(result)-1].branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "bare":
			result[len(result)-1].bare = true
		}
	}
	return result
}

type worktrees struct {
	list    []worktree
	current string // top-level directory of the current worktree
}

func loadWorktrees() (*worktrees, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil, fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return &worktrees{
		list:    parseWorktrees(stdout.String()),
		current: strings.TrimSpace(string(out)),
	}, nil
}

// pathOf returns the worktree that has branch checked out, or "".
func (w *worktrees) pathOf(branch string) string {
	for _, wt := range w.list {
		if wt.branch == branch {
			return wt.path
		}
	}
	return ""
}

// linkedMain returns the main worktree's path when the current directory is
// in a linked worktree, or "" otherwise. Branches are switched only in the
// main worktree so that each linked worktree keeps its own branch. A bare
// main repository has no working tree to switch in, so it reports "".
func (w *worktrees) linkedMain() string {
	if len(w.list) == 0 || w.list[0].bare || samePath(w.list[0].path, w.current) {
		return ""
	}
	return w.list[0].path
}

func samePath(a, b string) bool {
	if ra, err := filepath.EvalSymlinks(a); err == nil {
		a = ra
	}
	if rb, err := filepath.EvalSymlinks(b); err == nil {
		b = rb
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

// switchToWorktree runs cmds inside the worktree at path and asks the shell
// wrapper to cd there. Without the wrapper it only tells the user where to go
// and how to set the wrapper up. reason explains why it moves.
func switchToWorktree(path, reason string, cmds [][]string) error {
	cwdFile := os.Getenv(cwdFileEnv)
	if cwdFile == "" {
		fmt.Fprintf(os.Stderr, "%s: %s\n\n%s", reason, path, initHint)
		return fmt.Errorf("not switched: shell integration is not set up")
	}

	if err := os.WriteFile(cwdFile, []byte(path), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", cwdFileEnv, err)
	}
	fmt.Fprintf(os.Stderr, "%s; moving to %s\n", reason, path)
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = path
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", strings.Join(args, " "), err)
		}
	}
	return nil
}

const initHint = `To move there automatically, add the shell integration:

  # ~/.zshrc
  eval "$(gh list-pr --init zsh)"
  # ~/.bashrc
  eval "$(gh list-pr --init bash)"
  # ~/.config/fish/config.fish
  gh list-pr --init fish | source
`
