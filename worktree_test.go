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

worktree /wt/gone
HEAD 4444444444444444444444444444444444444444
branch refs/heads/gone
prunable gitdir file points to non-existent location
`
	got := parseWorktrees(out)
	want := []worktree{
		{path: "/repo", branch: "main"},
		{path: "/wt/feature", branch: "feature/x"},
		{path: "/wt/detached"},
		{path: "/wt/bare", bare: true},
		{path: "/wt/gone", branch: "gone", prunable: true},
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

func TestWorktreeRows(t *testing.T) {
	w := &worktrees{
		list: []worktree{
			{path: "/repo", branch: "topic"},
			{path: "/wt/pr", branch: "feature/pr"},
			{path: "/wt/main", branch: "main"},
			{path: "/wt/nopr", branch: "experiment"},
			{path: "/wt/detached"},
			{path: "/wt/gone", branch: "gone", prunable: true},
			{path: "/wt/current", branch: "current"},
		},
		current: "/wt/current",
	}
	prs := []PullRequest{
		{Number: 1, HeadRefName: "feature/pr"},
		{HeadRefName: "main"},
	}

	got := worktreeRows(prs, w)
	want := []struct{ path, branch string }{
		{"/repo", "topic"},
		{"/wt/nopr", "experiment"},
		{"/wt/detached", "(detached)"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows %+v, want %d", len(got), got, len(want))
	}
	for i, w := range want {
		if got[i].WorktreePath != w.path || got[i].HeadRefName != w.branch {
			t.Errorf("%d: got (%q, %q), want (%q, %q)", i, got[i].WorktreePath, got[i].HeadRefName, w.path, w.branch)
		}
		if got[i].Number != 0 {
			t.Errorf("%d: Number = %d, want 0", i, got[i].Number)
		}
	}
}

func TestAbbrevHome(t *testing.T) {
	tests := []struct{ path, home, want string }{
		{"/home/u/wt/foo", "/home/u", "~/wt/foo"},
		{"/home/u", "/home/u", "~"},
		{"/home/user2/wt", "/home/u", "/home/user2/wt"},
		{"/srv/wt", "", "/srv/wt"},
	}
	for _, tt := range tests {
		if got := abbrevHome(tt.path, tt.home); got != tt.want {
			t.Errorf("abbrevHome(%q, %q) = %q, want %q", tt.path, tt.home, got, tt.want)
		}
	}
}
