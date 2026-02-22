package main

import (
	"strings"
	"testing"
)

func TestBuildLine(t *testing.T) {
	baseLayout := ColumnLayout{
		NumWidth:    4,
		AddWidth:    3,
		DelWidth:    3,
		FileWidth:   2,
		TitleWidth:  20,
		AuthorWidth: 10,
		HeadRefWidth: 15,
		ShowFiles:   true,
		ShowDate:    true,
		ShowTitle:   true,
		ShowAuthor:  true,
	}

	basePR := PullRequest{
		Number:       42,
		Title:        "Fix critical bug",
		HeadRefName:  "feature-branch",
		AuthorName:   "alice",
		CreatedAt:    "2025-01-15T10:00:00Z",
		IsDraft:      false,
		Additions:    10,
		Deletions:    5,
		ChangedFiles: 3,
	}

	t.Run("normal_pr", func(t *testing.T) {
		got := buildLine(basePR, baseLayout)
		for _, want := range []string{"#42", "alice", "Fix critical bug", "feature-branch", "+", "10", "-", "5"} {
			if !strings.Contains(got, want) {
				t.Errorf("buildLine() missing %q in output: %q", want, got)
			}
		}
		if !strings.Contains(got, green) {
			t.Error("buildLine() normal PR should use green color for number")
		}
	})

	t.Run("draft_pr", func(t *testing.T) {
		draft := basePR
		draft.IsDraft = true
		got := buildLine(draft, baseLayout)
		if !strings.Contains(got, brightBlack) {
			t.Error("buildLine() draft PR should use brightBlack color for number")
		}
	})

	t.Run("hide_title_and_author", func(t *testing.T) {
		layout := baseLayout
		layout.ShowTitle = false
		layout.ShowAuthor = false
		got := buildLine(basePR, layout)
		if strings.Contains(got, "Fix critical bug") {
			t.Error("buildLine() with ShowTitle=false should not contain the title")
		}
		if strings.Contains(got, magenta) {
			t.Error("buildLine() with ShowAuthor=false should not contain author color")
		}
	})

	t.Run("show_files_and_date", func(t *testing.T) {
		got := buildLine(basePR, baseLayout)
		if !strings.Contains(got, "files") {
			t.Error("buildLine() with ShowFiles=true should contain 'files'")
		}
		if !strings.Contains(got, "2025-01-15T10:00:00Z") {
			t.Error("buildLine() with ShowDate=true should contain the date")
		}
	})
}

func TestFormatLines(t *testing.T) {
	layout := ColumnLayout{
		NumWidth:     4,
		AddWidth:     2,
		DelWidth:     2,
		FileWidth:    1,
		TitleWidth:   10,
		AuthorWidth:  5,
		HeadRefWidth: 10,
		ShowFiles:    false,
		ShowDate:     false,
		ShowTitle:    true,
		ShowAuthor:   true,
	}

	t.Run("empty_slice", func(t *testing.T) {
		got := formatLines(nil, layout)
		if got != "" {
			t.Errorf("formatLines(nil) = %q, want empty", got)
		}
	})

	t.Run("single_pr", func(t *testing.T) {
		prs := []PullRequest{
			{Number: 1, AuthorName: "a", Title: "t", HeadRefName: "b", Additions: 1, Deletions: 0},
		}
		got := formatLines(prs, layout)
		if !strings.HasSuffix(got, "\n") {
			t.Error("formatLines() single PR should end with newline")
		}
		if count := strings.Count(got, "\n"); count != 1 {
			t.Errorf("formatLines() single PR has %d newlines, want 1", count)
		}
	})

	t.Run("multiple_prs", func(t *testing.T) {
		prs := []PullRequest{
			{Number: 1, AuthorName: "a", Title: "t1", HeadRefName: "b1", Additions: 1, Deletions: 0},
			{Number: 2, AuthorName: "b", Title: "t2", HeadRefName: "b2", Additions: 2, Deletions: 1},
			{Number: 3, AuthorName: "c", Title: "t3", HeadRefName: "b3", Additions: 0, Deletions: 3},
		}
		got := formatLines(prs, layout)
		if count := strings.Count(got, "\n"); count != len(prs) {
			t.Errorf("formatLines() line count = %d, want %d", count, len(prs))
		}
	})
}
