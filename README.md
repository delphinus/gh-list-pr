# gh-list-pr

A GitHub CLI extension to list pull requests and interactively select one to checkout using fzf.

## Installation

```bash
gh extension install delphinus/gh-list-pr
```

## Usage

```bash
# Launch fzf and choose a PR to checkout
gh list-pr

# Print all active PRs
gh list-pr -p

# Open selected PR in browser
gh list-pr -w

# Filter PRs
gh list-pr -s '--author=@me'

# Show more PRs (default: 30)
gh list-pr -s '--limit 100'

# Include closed/merged PRs (default: open only)
gh list-pr -s '--state all'

# Custom fzf options
gh list-pr -f '--height=50%'
```

## Options

| Flag | Description |
|------|-------------|
| `-p`, `--print` | Print list without launching fzf |
| `-s`, `--search-options` | Filter PRs (passed to `gh pr list`). Note: `gh pr list` defaults to **30 items** and **open state only**. Use `--limit` and `--state` to override. |
| `-w`, `--web` | Open selected PR in web browser |
| `-f`, `--fzf-options` | Additional fzf options |

## Features

- Color-coded PR list with author, title, branch, additions/deletions, changed files, and date
- GitHub emoji support in PR titles (`:emoji_name:` → Unicode)
- Smart column layout with priority-based truncation for narrow terminals
- Default branch display (main/master/develop/staging)
- East Asian wide character support

## Requirements

- `git`
- `gh` (GitHub CLI)
- `fzf` (optional, falls back to print mode)
