package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/pflag"
)

type options struct {
	print         bool
	searchOptions string
	web           bool
	fzfOptions    string
}

func main() {
	var opt options
	pflag.BoolVarP(&opt.print, "print", "p", false, "Print list without launching fzf selector")
	pflag.StringVarP(&opt.searchOptions, "search-options", "s", "", "Filter PRs (passed to gh pr list)")
	pflag.BoolVarP(&opt.web, "web", "w", false, "Open selected PR in web browser")
	pflag.StringVarP(&opt.fzfOptions, "fzf-options", "f", "", "Additional fzf options")
	pflag.Parse()

	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintln(os.Stderr, "git not found")
		os.Exit(2)
	}
	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Fprintln(os.Stderr, "gh not found")
		os.Exit(2)
	}
	if !opt.print {
		if _, err := exec.LookPath("fzf"); err != nil {
			fmt.Fprintln(os.Stderr, "fzf not found")
			opt.print = true
		}
	}

	prs, err := fetchPRs(opt.searchOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch PRs: %v\n", err)
		os.Exit(1)
	}

	if opt.searchOptions == "" {
		branches, err := defaultBranches()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get default branches: %v\n", err)
		} else {
			prs = append(prs, branches...)
		}
	}

	emoji, err := loadEmoji()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load emojis: %v\n", err)
		os.Exit(1)
	}

	for i := range prs {
		prs[i].Title = replaceEmoji(prs[i].Title, emoji)
		if prs[i].Author.Login != "" {
			prs[i].AuthorName = prs[i].Author.Login
		} else {
			prs[i].AuthorName = "unknown"
		}
	}

	layout := calculateLayout(prs, opt)
	lines := formatLines(prs, layout)

	if opt.print {
		fmt.Print(lines)
		return
	}

	if err := runFzf(lines, opt); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
