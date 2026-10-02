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

// parseWorktrees parses `git worktree list --porcelain` and returns a map from
// local branch name to worktree path.
func parseWorktrees(out string) map[string]string {
	result := map[string]string{}
	var path string
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch refs/heads/"):
			if path != "" {
				result[strings.TrimPrefix(line, "branch refs/heads/")] = path
			}
		case line == "":
			path = ""
		}
	}
	return result
}

// findWorktree returns the path of another worktree that has branch checked
// out, or "" if there is none. The current worktree is not reported.
func findWorktree(branch string) (string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git worktree list: %w", err)
	}
	path, ok := parseWorktrees(stdout.String())[branch]
	if !ok {
		return "", nil
	}

	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	if samePath(path, strings.TrimSpace(string(out))) {
		return "", nil
	}
	return path, nil
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

// switchToWorktree updates the branch inside the worktree at path and asks the
// shell wrapper to cd there. Without the wrapper it only tells the user where
// the branch is and how to set the wrapper up.
func switchToWorktree(path, branch string, update [][]string) error {
	cwdFile := os.Getenv(cwdFileEnv)
	if cwdFile == "" {
		fmt.Fprintf(os.Stderr, "%s is checked out in another worktree: %s\n\n%s", branch, path, initHint)
		return fmt.Errorf("not switched: shell integration is not set up")
	}

	if err := os.WriteFile(cwdFile, []byte(path), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", cwdFileEnv, err)
	}
	fmt.Fprintf(os.Stderr, "%s is checked out in another worktree; moving to %s\n", branch, path)
	for _, args := range update {
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
