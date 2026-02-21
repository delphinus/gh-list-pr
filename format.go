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
)

func buildLine(pr PullRequest, layout ColumnLayout) string {
	var b strings.Builder

	// PR number
	numColor := green
	if pr.IsDraft {
		numColor = brightBlack
	}
	fmt.Fprintf(&b, "%s#%-*d  %s", numColor, layout.NumWidth, pr.Number, reset)

	// Author
	if layout.ShowAuthor {
		fmt.Fprintf(&b, "%s%s  %s", magenta, truncatePad(pr.AuthorName, layout.AuthorWidth), reset)
	}

	// Title
	if layout.ShowTitle {
		fmt.Fprintf(&b, "%s  ", truncatePad(pr.Title, layout.TitleWidth))
	}

	// Branch
	fmt.Fprintf(&b, "%s%s  %s", cyan, truncatePad(pr.HeadRefName, layout.HeadRefWidth), reset)

	// Additions/Deletions
	fmt.Fprintf(&b, "%s+%*d%s/%s-%*d%s",
		green, layout.AddWidth, pr.Additions, reset,
		red, layout.DelWidth, pr.Deletions, reset)

	// Changed files
	if layout.ShowFiles {
		fmt.Fprintf(&b, "  %*d files", layout.FileWidth, pr.ChangedFiles)
	}

	// Date
	if layout.ShowDate {
		fmt.Fprintf(&b, "  %s%s%s", brightBlack, pr.CreatedAt, reset)
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
