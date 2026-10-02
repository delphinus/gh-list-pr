# gh-list-pr

A GitHub CLI extension to list pull requests and interactively select one to checkout using fzf.

![demo](demo.gif)

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

# Print shell integration (moves to the worktree if the branch is checked out there)
gh list-pr --init zsh
```

## Options

| Flag | Description |
|------|-------------|
| `-p`, `--print` | Print list without launching fzf |
| `-s`, `--search-options` | Filter PRs (passed to `gh pr list`). Note: `gh pr list` defaults to **30 items** and **open state only**. Use `--limit` and `--state` to override. |
| `-w`, `--web` | Open selected PR in web browser |
| `-f`, `--fzf-options` | Additional fzf options |
| `--init` | Print shell integration for `zsh`, `bash` or `fish` (see [Worktrees](#worktrees)) |
| `--cmd` | Function name defined by `--init` (default: `gh`, which wraps `gh` itself) |

## Worktrees

If the selected branch is already checked out in another [worktree](https://git-scm.com/docs/git-worktree), `gh list-pr` moves your shell there and updates the branch in that worktree, instead of failing with `already used by worktree`.

Branches are switched only in the main worktree, so each linked worktree keeps the branch it was made for. Inside a linked worktree:

| Selected branch | What happens |
|---|---|
| The branch of this worktree | Updated in place |
| Checked out in another worktree (including the main one) | Moves there and updates it |
| Not checked out anywhere | Moves to the main worktree and switches branches there |

`-b` is refused inside a linked worktree for the same reason. If the main repository is bare, branches are switched in place as before.

A program cannot change its parent shell's directory, so this needs the shell integration:

```bash
# ~/.zshrc
eval "$(gh list-pr --init zsh)"

# ~/.bashrc
eval "$(gh list-pr --init bash)"

# ~/.config/fish/config.fish
gh list-pr --init fish | source
```

This defines a `gh` function that handles only `gh list-pr` and passes every other subcommand through to the real `gh`, so you keep typing `gh list-pr` and `gh` completion keeps working. The worktree path is handed over through a temporary file named by `GH_LIST_PR_CWD_FILE`, never through stdout.

- If `gh` is already a function or alias in your shell, the integration does nothing and prints a warning. Use `--cmd` to define a separate function instead, e.g. `eval "$(gh list-pr --init zsh --cmd glp)"` and run `glp`.
- zsh and bash use `builtin cd`, so a `cd` replaced by an alias (e.g. `zoxide init --cmd cd`) does not interfere. fish uses fish's own `cd` so that `cd -`, `prevd` and `cdh` still remember where you came from.
- Without the integration, `gh list-pr` leaves the branch untouched, prints the worktree path and how to set the integration up, and exits with status 1.

## Features

- Color-coded PR list with author, title, branch, additions/deletions, changed files, and date
- GitHub emoji support in PR titles (`:emoji_name:` → Unicode)
- Smart column layout with priority-based truncation for narrow terminals
- Default branch display (main/master/develop/staging)
- East Asian wide character support

## Why not `gh pr checkout`?

`gh pr checkout` (without arguments) also offers interactive PR selection, but `gh list-pr` provides a richer experience:

| | `gh pr checkout` | `gh list-pr` |
|---|---|---|
| Selector | Built-in simple picker | fzf (incremental search) |
| Max items | 10 (fixed) | Configurable (`-s '--limit 1000'`) |
| Displayed info | Number, title, branch | Author, title, branch, +/-lines, changed files, date |
| Color coding | Minimal | Full (additions in green, deletions in red, etc.) |
| Filtering | None | Via `gh pr list` options (`--author`, `--state`, `--search`, etc.) |
| Default branches | Not shown | Shown (main/master/develop/staging) |
| Output modes | Interactive only | Interactive, print (`-p`), web (`-w`) |
| fzf customization | N/A | `--border`, `--height`, `--padding`, etc. via `-f` |

## Requirements

- `git`
- `gh` (GitHub CLI)
- `fzf` (optional, falls back to print mode)
