package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestParseWorktrees(t *testing.T) {
	out := `worktree /repo
HEAD 1111111111111111111111111111111111111111
branch refs/heads/main

worktree /wt/feature
HEAD 2222222222222222222222222222222222222222
branch refs/heads/feature/x

worktree /wt/detached
HEAD 3333333333333333333333333333333333333333
detached

worktree /wt/bare
bare
`
	got := parseWorktrees(out)
	want := []worktree{
		{path: "/repo", branch: "main"},
		{path: "/wt/feature", branch: "feature/x"},
		{path: "/wt/detached"},
		{path: "/wt/bare", bare: true},
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestWorktrees(t *testing.T) {
	list := []worktree{
		{path: "/repo", branch: "main"},
		{path: "/wt/feature", branch: "feature/x"},
	}

	inMain := &worktrees{list: list, current: "/repo"}
	if got := inMain.linkedMain(); got != "" {
		t.Errorf("in main worktree: linkedMain() = %q, want empty", got)
	}

	inLinked := &worktrees{list: list, current: "/wt/feature"}
	if got := inLinked.linkedMain(); got != "/repo" {
		t.Errorf("in linked worktree: linkedMain() = %q, want /repo", got)
	}
	if got := inLinked.pathOf("feature/x"); got != "/wt/feature" {
		t.Errorf("pathOf(feature/x) = %q", got)
	}
	if got := inLinked.pathOf("other"); got != "" {
		t.Errorf("pathOf(other) = %q, want empty", got)
	}

	bare := &worktrees{list: []worktree{{path: "/repo.git", bare: true}, list[1]}, current: "/wt/feature"}
	if got := bare.linkedMain(); got != "" {
		t.Errorf("bare main: linkedMain() = %q, want empty", got)
	}
}

func TestShellInit(t *testing.T) {
	for _, shell := range []string{"zsh", "bash", "fish"} {
		for _, cmd := range []string{"gh", "glp"} {
			t.Run(shell+"/"+cmd, func(t *testing.T) {
				script, err := shellInit(shell, cmd)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(script, cwdFileEnv) {
					t.Errorf("script does not set %s", cwdFileEnv)
				}
				if strings.Contains(script, "{{") {
					t.Errorf("unreplaced placeholder:\n%s", script)
				}
				if cmd == "gh" && !strings.Contains(script, "list-pr") {
					t.Errorf("gh wrapper does not dispatch on list-pr")
				}

				bin, err := exec.LookPath(shell)
				if err != nil {
					t.Skipf("%s not found", shell)
				}
				c := exec.Command(bin, "-n")
				c.Stdin = strings.NewReader(script)
				if out, err := c.CombinedOutput(); err != nil {
					t.Errorf("syntax error: %v\n%s\n%s", err, out, script)
				}
			})
		}
	}
}

func TestShellInitErrors(t *testing.T) {
	if _, err := shellInit("tcsh", "gh"); err == nil {
		t.Error("expected error for unsupported shell")
	}
	if _, err := shellInit("zsh", "a;b"); err == nil {
		t.Error("expected error for invalid command name")
	}
}
