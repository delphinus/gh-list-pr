package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

var selectionRe = regexp.MustCompile(`^#(\d+).*\s+(\S+)\s+\+\d+/-\d+`)

func runFzf(lines string, opt options) error {
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
	return handleSelection(selected, opt)
}

func handleSelection(selected string, opt options) error {
	m := selectionRe.FindStringSubmatch(selected)
	if m == nil {
		return fmt.Errorf("failed to parse selection: %s", selected)
	}

	num, _ := strconv.Atoi(m[1])
	ref := m[2]

	if num == 0 {
		return execCommand("sh", "-c",
			fmt.Sprintf("git checkout %s && git pull origin %s && git submodule update --init --recursive", ref, ref))
	}
	if opt.web {
		return execCommand("gh", "pr", "view", "-w", m[1])
	}
	return execCommand("gh", "co", "--recurse-submodules", m[1])
}

func execCommand(name string, args ...string) error {
	bin, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s not found: %w", name, err)
	}
	argv := append([]string{name}, args...)
	return syscall.Exec(bin, argv, os.Environ())
}
