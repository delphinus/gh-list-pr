package main

import (
	"testing"
)

func TestSelectionRegex(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMatch  bool
		wantNum    string
		wantBranch string
	}{
		{
			name:       "standard_pr",
			input:      "#42  user  Fix bug  feature-branch  +10/-5",
			wantMatch:  true,
			wantNum:    "42",
			wantBranch: "feature-branch",
		},
		{
			name:       "default_branch",
			input:      "#0  system  main  main  +0/-0",
			wantMatch:  true,
			wantNum:    "0",
			wantBranch: "main",
		},
		{
			name:       "large_pr_number",
			input:      "#12345  user  Some title  my-branch  +100/-200",
			wantMatch:  true,
			wantNum:    "12345",
			wantBranch: "my-branch",
		},
		{
			name:      "empty_string",
			input:     "",
			wantMatch: false,
		},
		{
			name:      "invalid_string",
			input:     "not a pr line",
			wantMatch: false,
		},
		{
			name:       "no_ansi",
			input:      "#7  user  Title  branch-name  +1/-0",
			wantMatch:  true,
			wantNum:    "7",
			wantBranch: "branch-name",
		},
		{
			name:       "padded_deletions",
			input:      "#34473  lewis6991  feat(lua): add vim.async  vimasync  +2545/-   0",
			wantMatch:  true,
			wantNum:    "34473",
			wantBranch: "vimasync",
		},
		{
			name:       "padded_additions",
			input:      "#100  user  Title  branch  +   5/-300",
			wantMatch:  true,
			wantNum:    "100",
			wantBranch: "branch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := selectionRe.FindStringSubmatch(tt.input)
			if tt.wantMatch {
				if m == nil {
					t.Fatalf("expected match for %q, got nil", tt.input)
				}
				if m[1] != tt.wantNum {
					t.Errorf("PR number = %q, want %q", m[1], tt.wantNum)
				}
				if m[2] != tt.wantBranch {
					t.Errorf("branch = %q, want %q", m[2], tt.wantBranch)
				}
			} else {
				if m != nil {
					t.Errorf("expected no match for %q, got %v", tt.input, m)
				}
			}
		})
	}
}
