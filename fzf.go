package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func switchBack() error {
	wts, err := loadWorktrees()
	if err != nil {
		return err
	}
	if main := wts.linkedMain(); main != "" {
		return fmt.Errorf("--back switches branches, which is done only in the main worktree: %s", main)
	}
	return runCommands([][]string{
		{"git", "checkout", "@{-1}"},
		{"git", "submodule", "update", "--init", "--recursive"},
	})
}

func runFzf(lines string, prs []PullRequest, opt options) error {
	args := []string{"--ansi"}

	// Merge user fzf options, avoiding duplicate --ansi
	if opt.fzfOptions != "" {
		extra := opt.fzfOptions
		if strings.Contains(extra, "--ansi") {
			extra = strings.ReplaceAll(extra, "--ansi", "")
		}
		if fields := strings.Fields(extra); len(fields) > 0 {
			args = append(args, fields...)
		}
	}

	cmd := exec.Command("fzf", args...)
	cmd.Stdin = strings.NewReader(lines)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cancelled")
	}

	selected := strings.TrimSpace(string(out))
	i := selectedIndex(lines, selected)
	if i < 0 || i >= len(prs) {
		return fmt.Errorf("failed to parse selection: %s", selected)
	}
	return handleSelection(prs[i], opt)
}

// selectedIndex returns the index of the line fzf printed for the selection,
// or -1. fzf prints the line with the colors removed. Columns may be
// truncated, so the row is identified by its whole line rather than parsed.
func selectedIndex(lines, selected string) int {
	selected = strings.TrimSpace(ansiRe.ReplaceAllString(selected, ""))
	for i, line := range strings.Split(strings.TrimSuffix(lines, "\n"), "\n") {
		if strings.TrimSpace(ansiRe.ReplaceAllString(line, "")) == selected {
			return i
		}
	}
	return -1
}

func handleSelection(pr PullRequest, opt options) error {
	if pr.WorktreePath != "" {
		if opt.web {
			return fmt.Errorf("no pull request for the worktree: %s", pr.WorktreePath)
		}
		return switchToWorktree(pr.WorktreePath, "selected a worktree", nil)
	}

	num := pr.Number
	ref := pr.HeadRefName
	numStr := strconv.Itoa(num)

	if opt.web {
		return execCommand("gh", "pr", "view", "-w", numStr)
	}

	cmds := [][]string{{"gh", "co", "--recurse-submodules", numStr}}
	if num == 0 {
		cmds = [][]string{
			{"git", "checkout", ref},
			{"git", "pull", "origin", ref},
			{"git", "submodule", "update", "--init", "--recursive"},
		}
	}

	wts, err := loadWorktrees()
	if err != nil {
		return err
	}
	if path := wts.pathOf(ref); path != "" {
		if !samePath(path, wts.current) {
			return switchToWorktree(path, ref+" is checked out in another worktree", cmds)
		}
	} else if main := wts.linkedMain(); main != "" {
		return switchToWorktree(main, "switching branches in the main worktree, not in this linked worktree", cmds)
	}

	if num == 0 {
		return runCommands(cmds)
	}
	return execCommand(cmds[0][0], cmds[0][1:]...)
}

func runCommands(cmds [][]string) error {
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", strings.Join(args, " "), err)
		}
	}
	return nil
}
