package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

func termWidth() int {
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if w, err := strconv.Atoi(cols); err == nil && w > 0 {
			return w
		}
	}

	if f, err := os.Open("/dev/tty"); err == nil {
		defer f.Close()
		if w, _, err := term.GetSize(int(f.Fd())); err == nil && w > 0 {
			return w
		}
	}

	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}

	return 80
}

var (
	borderRe  = regexp.MustCompile(`--border(?:=(\S+))?`)
	paddingRe = regexp.MustCompile(`--padding[= ]'?([^'"\s]+)'?`)
)

func fzfMargin(opt options) int {
	if opt.print {
		return 0
	}
	allOpts := os.Getenv("FZF_DEFAULT_OPTS") + " " + opt.fzfOptions
	margin := 2 // fzf pointer/indicator

	// --border detection (last match wins)
	style := ""
	for _, m := range borderRe.FindAllStringSubmatch(allOpts, -1) {
		if m[1] != "" {
			style = m[1]
		} else {
			style = "rounded"
		}
	}
	switch style {
	case "rounded", "sharp", "bold", "double", "block", "thinblock", "vertical":
		margin += 2
	case "left", "right":
		margin += 1
	}

	// --padding detection (add left+right)
	if m := paddingRe.FindStringSubmatch(allOpts); m != nil {
		parts := strings.Split(m[1], ",")
		switch len(parts) {
		case 1:
			if v, err := strconv.Atoi(parts[0]); err == nil {
				margin += v * 2
			}
		case 2:
			if v, err := strconv.Atoi(parts[1]); err == nil {
				margin += v * 2
			}
		default:
			if len(parts) >= 4 {
				v1, e1 := strconv.Atoi(parts[1])
				v3, e3 := strconv.Atoi(parts[3])
				if e1 == nil && e3 == nil {
					margin += v1 + v3
				}
			}
		}
	}

	return margin
}

func displayWidth(s string) int {
	return runewidth.StringWidth(s)
}

func truncatePad(s string, width int) string {
	if width <= 0 {
		return ""
	}
	dw := runewidth.StringWidth(s)
	if dw <= width {
		return s + strings.Repeat(" ", width-dw)
	}
	return runewidth.Truncate(s, width, "\u2026")
}
