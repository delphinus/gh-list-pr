package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

var snapshotPRs = []PullRequest{
	{
		Number:       1,
		Title:        "Add new feature",
		HeadRefName:  "feature/add-new",
		AuthorName:   "alice",
		CreatedAt:    "2025-01-15T10:00:00Z",
		IsDraft:      false,
		Additions:    42,
		Deletions:    10,
		ChangedFiles: 5,
	},
	{
		Number:       23,
		Title:        "Draft: WIP refactor",
		HeadRefName:  "refactor/cleanup",
		AuthorName:   "bob",
		CreatedAt:    "2025-01-16T12:00:00Z",
		IsDraft:      true,
		Additions:    150,
		Deletions:    200,
		ChangedFiles: 15,
	},
	{
		Number:       456,
		Title:        "日本語のタイトル",
		HeadRefName:  "fix/i18n-support",
		AuthorName:   "charlie",
		CreatedAt:    "2025-01-17T14:00:00Z",
		IsDraft:      false,
		Additions:    5,
		Deletions:    3,
		ChangedFiles: 2,
	},
	{
		Number:       7890,
		Title:        "Big changes everywhere",
		HeadRefName:  "release/v2.0",
		AuthorName:   "dave",
		CreatedAt:    "2025-01-18T16:00:00Z",
		IsDraft:      false,
		Additions:    1234,
		Deletions:    567,
		ChangedFiles: 89,
	},
}

func TestSnapshot(t *testing.T) {
	t.Setenv("FZF_DEFAULT_OPTS", "")

	tests := []struct {
		name   string
		cols   string
		golden string
	}{
		{"wide", "120", "testdata/snapshot_wide.golden"},
		{"medium", "80", "testdata/snapshot_medium.golden"},
		{"narrow", "50", "testdata/snapshot_narrow.golden"},
		{"minimal", "30", "testdata/snapshot_minimal.golden"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("COLUMNS", tt.cols)
			layout := calculateLayout(snapshotPRs, options{print: true})
			got := formatLines(snapshotPRs, layout)

			if *update {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatalf("create testdata dir: %v", err)
				}
				if err := os.WriteFile(tt.golden, []byte(got), 0o644); err != nil {
					t.Fatalf("write golden file: %v", err)
				}
				return
			}

			want, err := os.ReadFile(tt.golden)
			if err != nil {
				t.Fatalf("read golden file: %v (run with -update to create)", err)
			}
			if got != string(want) {
				t.Errorf("snapshot mismatch:\n%s", lineDiff(string(want), got))
			}
		})
	}
}

func lineDiff(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")
	var b strings.Builder
	max := len(wantLines)
	if len(gotLines) > max {
		max = len(gotLines)
	}
	for i := 0; i < max; i++ {
		var wl, gl string
		if i < len(wantLines) {
			wl = wantLines[i]
		}
		if i < len(gotLines) {
			gl = gotLines[i]
		}
		if wl != gl {
			fmt.Fprintf(&b, "line %d:\n  want: %q\n  got:  %q\n", i+1, wl, gl)
		}
	}
	return b.String()
}
