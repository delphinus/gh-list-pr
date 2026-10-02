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
	want := map[string]string{
		"main":      "/repo",
		"feature/x": "/wt/feature",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %q, want %q", k, got[k], v)
		}
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
