package main

import (
	"fmt"
	"regexp"
	"strings"
)

var cmdNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

// shellInit returns the shell integration script. With cmd "gh" it wraps gh
// itself and intercepts only `gh list-pr`; otherwise it defines cmd as a
// shortcut for `gh list-pr`.
func shellInit(shell, cmd string) (string, error) {
	if !cmdNameRe.MatchString(cmd) {
		return "", fmt.Errorf("invalid command name: %q", cmd)
	}
	var tmpl string
	switch shell {
	case "zsh":
		tmpl = zshInit
	case "bash":
		tmpl = bashInit
	case "fish":
		tmpl = fishInit
	default:
		return "", fmt.Errorf("unsupported shell: %q (zsh, bash, fish)", shell)
	}

	// pre: code before running gh list-pr; args: arguments passed to it.
	pre, args := "", `list-pr "$@"`
	if shell == "fish" {
		args = "list-pr $argv"
	}
	if cmd == "gh" {
		switch shell {
		case "fish":
			pre = "    if test \"$argv[1]\" != list-pr\n        command gh $argv\n        return\n    end\n"
			args = "$argv"
		default:
			pre = "  if [ \"$1\" != list-pr ]; then\n    command gh \"$@\"\n    return\n  fi\n"
			args = `"$@"`
		}
	}

	r := strings.NewReplacer("{{cmd}}", cmd, "{{pre}}", pre, "{{args}}", args, "{{env}}", cwdFileEnv)
	return r.Replace(tmpl), nil
}

// The wrapper is skipped when cmd is already a function or alias defined by
// someone else; redefining it on re-sourcing the rc file is fine.

const zshInit = `# gh-list-pr shell integration for zsh
if (( $+aliases[{{cmd}}] )) || { (( $+functions[{{cmd}}] )) && [[ $functions[{{cmd}}] != *{{env}}* ]] }; then
  print -u2 'gh-list-pr: {{cmd}} is already a function or alias; skipped. Try: eval "$(gh list-pr --init zsh --cmd glp)"'
else
function {{cmd}} {
{{pre}}  local cwd_file ret
  cwd_file=$(command mktemp) || return
  {{env}}=$cwd_file command gh {{args}}
  ret=$?
  if [ -s "$cwd_file" ]; then
    builtin cd -- "$(<"$cwd_file")" || ret=$?
  fi
  command rm -f -- "$cwd_file"
  return $ret
}
fi
`

const bashInit = `# gh-list-pr shell integration for bash
if alias {{cmd}} >/dev/null 2>&1 || { declare -F {{cmd}} >/dev/null && ! declare -f {{cmd}} | command grep -q {{env}}; }; then
  echo 'gh-list-pr: {{cmd}} is already a function or alias; skipped. Try: eval "$(gh list-pr --init bash --cmd glp)"' >&2
else
function {{cmd}} {
{{pre}}  local cwd_file ret
  cwd_file=$(command mktemp) || return
  {{env}}=$cwd_file command gh {{args}}
  ret=$?
  if [ -s "$cwd_file" ]; then
    builtin cd -- "$(<"$cwd_file")" || ret=$?
  fi
  command rm -f -- "$cwd_file"
  return $ret
}
fi
`

// fish uses cd (not builtin cd) so that fish's own cd function records the
// directory history used by cd -, prevd and cdh.
const fishInit = `# gh-list-pr shell integration for fish
if functions -q {{cmd}}; and not functions {{cmd}} | string match -q '*{{env}}*'
    echo 'gh-list-pr: {{cmd}} is already a function; skipped. Try: gh list-pr --init fish --cmd glp | source' >&2
else
function {{cmd}}
{{pre}}    set -l cwd_file (command mktemp); or return
    {{env}}=$cwd_file command gh {{args}}
    set -l ret $status
    if test -s $cwd_file
        cd -- (command cat -- $cwd_file); or set ret $status
    end
    command rm -f -- $cwd_file
    return $ret
end
end
`
