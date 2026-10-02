package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var selectionRe = regexp.MustCompile(`^#(\d+).*\s+(\S+)\s+\+\s*\d+/-\s*\d+`)

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
	return handleSelection(selected, prs, opt)
}

func handleSelection(selected string, prs []PullRequest, opt options) error {
	m := selectionRe.FindStringSubmatch(selected)
	if m == nil {
		return fmt.Errorf("failed to parse selection: %s", selected)
	}

	num, _ := strconv.Atoi(m[1])
	ref := m[2]
	// The branch column may be truncated, so take the PR's branch from the data.
	if num != 0 {
		for _, pr := range prs {
			if pr.Number == num {
				ref = pr.HeadRefName
				break
			}
		}
	}

	if opt.web {
		return execCommand("gh", "pr", "view", "-w", m[1])
	}

	cmds := [][]string{{"gh", "co", "--recurse-submodules", m[1]}}
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
