package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/spf13/pflag"
)

type options struct {
	print         bool
	searchOptions string
	web           bool
	fzfOptions    string
	version       bool
}

func main() {
	var opt options
	pflag.BoolVarP(&opt.print, "print", "p", false, "Print list without launching fzf selector")
	pflag.StringVarP(&opt.searchOptions, "search-options", "s", "", "Filter PRs (passed to gh pr list; defaults to 30 items, open only)")
	pflag.BoolVarP(&opt.web, "web", "w", false, "Open selected PR in web browser")
	pflag.StringVarP(&opt.fzfOptions, "fzf-options", "f", "", "Additional fzf options")
	pflag.BoolVarP(&opt.version, "version", "v", false, "Print version")

	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr, `List pull requests and interactively select one to checkout using fzf.

Shows a color-coded PR list with author, title, branch, additions/deletions,
changed files, and date. Default branches (main/master/develop/staging) are
included when no search filter is applied.

USAGE
  gh list-pr [flags]

EXAMPLES
  # Launch fzf and choose a PR to checkout
  gh list-pr

  # Print all active PRs without fzf
  gh list-pr -p

  # Open selected PR in web browser
  gh list-pr -w

  # Filter PRs by author
  gh list-pr -s '--author=@me'

  # Show more PRs (default: 30)
  gh list-pr -s '--limit 100'

  # Include closed/merged PRs (default: open only)
  gh list-pr -s '--state all'

FLAGS`)
		pflag.PrintDefaults()
	}

	pflag.Parse()

	if opt.version {
		version := "(devel)"
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
			version = info.Main.Version
		}
		fmt.Println("gh-list-pr " + version)
		return
	}

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

	sp := newSpinner("Fetching pull requests...")
	sp.start()

	prs, err := fetchPRs(opt.searchOptions)
	if err != nil {
		sp.stop()
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
		sp.stop()
		fmt.Fprintf(os.Stderr, "Failed to load emojis: %v\n", err)
		os.Exit(1)
	}

	sp.stop()

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
