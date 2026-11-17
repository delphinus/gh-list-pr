package main

import (
	"testing"
)

func TestSelectedIndex(t *testing.T) {
	lines := "\033[32m#42  \033[0m\033[35muser  \033[0mFix bug  \033[36mfeature-…  \033[0m+10/-5\n" +
		"\033[32m#0   \033[0m\033[35msystem  \033[0mmain  \033[36mmain  \033[0m+0/-0\n" +
		"\033[34mwt   \033[0m\033[35mworktree  \033[0m~/wt/foo  \033[36mfoo  \033[0m+0/-0\n"

	tests := []struct {
		name     string
		selected string
		want     int
	}{
		{"pr", "#42  user  Fix bug  feature-…  +10/-5", 0},
		{"default_branch", "#0   system  main  main  +0/-0", 1},
		{"worktree", "wt   worktree  ~/wt/foo  foo  +0/-0", 2},
		{"surrounding_space", "  wt   worktree  ~/wt/foo  foo  +0/-0\n", 2},
		{"with_colors", "\033[34mwt   \033[0mworktree  ~/wt/foo  foo  +0/-0", 2},
		{"unknown", "#7  user  Title  branch  +1/-0", -1},
		{"empty", "", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectedIndex(lines, tt.selected); got != tt.want {
				t.Errorf("selectedIndex(%q) = %d, want %d", tt.selected, got, tt.want)
			}
		})
	}
}
