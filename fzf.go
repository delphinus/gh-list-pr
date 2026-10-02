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
	for _, args := range [][]string{
		{"git", "checkout", "@{-1}"},
		{"git", "submodule", "update", "--init", "--recursive"},
	} {
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

	if !opt.web {
		path, err := findWorktree(ref)
		if err != nil {
			return err
		}
		if path != "" {
			update := [][]string{
				{"gh", "co", "--recurse-submodules", m[1]},
			}
			if num == 0 {
				update = [][]string{
					{"git", "pull", "origin", ref},
					{"git", "submodule", "update", "--init", "--recursive"},
				}
			}
			return switchToWorktree(path, ref, update)
		}
	}

	if num == 0 {
		for _, args := range [][]string{
			{"git", "checkout", ref},
			{"git", "pull", "origin", ref},
			{"git", "submodule", "update", "--init", "--recursive"},
		} {
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("%s: %w", args[0], err)
			}
		}
		return nil
	}
	if opt.web {
		return execCommand("gh", "pr", "view", "-w", m[1])
	}
	return execCommand("gh", "co", "--recurse-submodules", m[1])
}
