package main

import (
	"testing"
)

func TestCalculateLayout(t *testing.T) {
	t.Run("empty_prs", func(t *testing.T) {
		t.Setenv("COLUMNS", "120")
		layout := calculateLayout(nil, options{print: true})
		if layout.NumWidth != 4 {
			t.Errorf("NumWidth = %d, want 4", layout.NumWidth)
		}
		if !layout.ShowTitle || !layout.ShowAuthor || !layout.ShowFiles || !layout.ShowDate {
			t.Error("empty PR list should show all columns")
		}
	})

	t.Run("wide_terminal", func(t *testing.T) {
		t.Setenv("COLUMNS", "200")
		prs := []PullRequest{
			{Number: 42, AuthorName: "alice", Title: "Fix bug", HeadRefName: "feature-branch",
				Additions: 10, Deletions: 5, ChangedFiles: 3},
		}
		layout := calculateLayout(prs, options{print: true})
		if !layout.ShowTitle || !layout.ShowAuthor || !layout.ShowFiles || !layout.ShowDate {
			t.Error("wide terminal should show all columns")
		}
		if layout.TitleWidth != displayWidth("Fix bug") {
			t.Errorf("TitleWidth = %d, want %d", layout.TitleWidth, displayWidth("Fix bug"))
		}
		if layout.AuthorWidth != displayWidth("alice") {
			t.Errorf("AuthorWidth = %d, want %d", layout.AuthorWidth, displayWidth("alice"))
		}
		if layout.HeadRefWidth != displayWidth("feature-branch") {
			t.Errorf("HeadRefWidth = %d, want %d", layout.HeadRefWidth, displayWidth("feature-branch"))
		}
	})

	t.Run("narrow_terminal", func(t *testing.T) {
		t.Setenv("COLUMNS", "40")
		prs := []PullRequest{
			{Number: 42, AuthorName: "alice-longname", Title: "This is a very long title for testing",
				HeadRefName: "feature/very-long-branch-name", Additions: 10, Deletions: 5, ChangedFiles: 3},
		}
		layout := calculateLayout(prs, options{print: true})
		// At 40 cols with long content, some columns must be hidden
		if layout.ShowFiles && layout.ShowDate {
			t.Error("narrow terminal should hide files or date")
		}
	})

	t.Run("large_pr_number", func(t *testing.T) {
		t.Setenv("COLUMNS", "200")
		prs := []PullRequest{
			{Number: 100000, AuthorName: "a", Title: "t", HeadRefName: "b",
				Additions: 1, Deletions: 1, ChangedFiles: 1},
		}
		layout := calculateLayout(prs, options{print: true})
		if layout.NumWidth != 6 {
			t.Errorf("NumWidth = %d, want 6 for PR #100000", layout.NumWidth)
		}
	})

	t.Run("large_additions_deletions", func(t *testing.T) {
		t.Setenv("COLUMNS", "200")
		prs := []PullRequest{
			{Number: 1, AuthorName: "a", Title: "t", HeadRefName: "b",
				Additions: 10000, Deletions: 10000, ChangedFiles: 1},
		}
		layout := calculateLayout(prs, options{print: true})
		if layout.AddWidth != 5 {
			t.Errorf("AddWidth = %d, want 5 for 10000 additions", layout.AddWidth)
		}
		if layout.DelWidth != 5 {
			t.Errorf("DelWidth = %d, want 5 for 10000 deletions", layout.DelWidth)
		}
	})

	t.Run("fullwidth_title", func(t *testing.T) {
		t.Setenv("COLUMNS", "200")
		prs := []PullRequest{
			{Number: 1, AuthorName: "a", Title: "日本語テスト", HeadRefName: "b",
				Additions: 1, Deletions: 1, ChangedFiles: 1},
		}
		layout := calculateLayout(prs, options{print: true})
		if layout.TitleWidth != displayWidth("日本語テスト") {
			t.Errorf("TitleWidth = %d, want %d for fullwidth title", layout.TitleWidth, displayWidth("日本語テスト"))
		}
	})

	t.Run("fzf_margin_reduces_width", func(t *testing.T) {
		t.Setenv("COLUMNS", "80")
		t.Setenv("FZF_DEFAULT_OPTS", "")
		prs := []PullRequest{
			{Number: 42, AuthorName: "alice-longname", Title: "This is a somewhat long title",
				HeadRefName: "feature/long-branch", Additions: 10, Deletions: 5, ChangedFiles: 3},
		}
		layoutPrint := calculateLayout(prs, options{print: true})
		layoutFzf := calculateLayout(prs, options{print: false, fzfOptions: "--border=rounded"})

		// fzf margin should reduce available width, potentially shrinking columns
		printTotal := layoutPrint.TitleWidth + layoutPrint.AuthorWidth + layoutPrint.HeadRefWidth
		fzfTotal := layoutFzf.TitleWidth + layoutFzf.AuthorWidth + layoutFzf.HeadRefWidth
		if fzfTotal > printTotal {
			t.Errorf("fzf layout variable width total (%d) should not exceed print layout (%d)",
				fzfTotal, printTotal)
		}
	})
}
