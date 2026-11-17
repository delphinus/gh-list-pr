package main

import (
	"fmt"
	"strings"
)

const (
	reset       = "\033[0m"
	green       = "\033[32m"
	red         = "\033[31m"
	cyan        = "\033[36m"
	magenta     = "\033[35m"
	brightBlack = "\033[90m"
	blue        = "\033[34m"
)

// worktreeMark replaces the PR number on rows for worktrees without a PR.
const worktreeMark = "wt"

func buildLine(pr PullRequest, layout ColumnLayout) string {
	var b strings.Builder

	isWT := pr.WorktreePath != ""

	// PR number
	if isWT {
		fmt.Fprintf(&b, "%s%-*s  %s", blue, layout.NumWidth+1, worktreeMark, reset)
	} else {
		numColor := green
		if pr.IsDraft {
			numColor = brightBlack
		}
		fmt.Fprintf(&b, "%s#%-*d  %s", numColor, layout.NumWidth, pr.Number, reset)
	}

	// Author
	if layout.ShowAuthor {
		fmt.Fprintf(&b, "%s%s  %s", magenta, truncatePad(pr.AuthorName, layout.AuthorWidth), reset)
	}

	// Title
	// The path of a worktree row takes the title column, in the color of its mark.
	if layout.ShowTitle {
		if isWT {
			fmt.Fprintf(&b, "%s%s  %s", blue, truncatePad(pr.Title, layout.TitleWidth), reset)
		} else {
			fmt.Fprintf(&b, "%s  ", truncatePad(pr.Title, layout.TitleWidth))
		}
	}

	// Branch
	fmt.Fprintf(&b, "%s%s  %s", cyan, truncatePad(pr.HeadRefName, layout.HeadRefWidth), reset)

	// Additions/Deletions and changed files. Worktree rows have no PR, so
	// their zeros are dimmed.
	if isWT {
		fmt.Fprintf(&b, "%s+%*d/-%*d", brightBlack, layout.AddWidth, pr.Additions, layout.DelWidth, pr.Deletions)
		if layout.ShowFiles {
			fmt.Fprintf(&b, "  %*d files", layout.FileWidth, pr.ChangedFiles)
		}
		b.WriteString(reset)
	} else {
		fmt.Fprintf(&b, "%s+%*d%s/%s-%*d%s",
			green, layout.AddWidth, pr.Additions, reset,
			red, layout.DelWidth, pr.Deletions, reset)
		if layout.ShowFiles {
			fmt.Fprintf(&b, "  %*d files", layout.FileWidth, pr.ChangedFiles)
		}
	}

	// Date
	if layout.ShowDate {
		fmt.Fprintf(&b, "  %s%s%s", pr.DateColor, pr.DateText, reset)
	}

	return b.String()
}

func formatLines(prs []PullRequest, layout ColumnLayout) string {
	var b strings.Builder
	for _, pr := range prs {
		b.WriteString(buildLine(pr, layout))
		b.WriteByte('\n')
	}
	return b.String()
}
