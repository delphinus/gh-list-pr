//go:build !windows

package main

import (
	"os"

	"golang.org/x/term"
)

func termWidthFromTTY() int {
	if f, err := os.Open("/dev/tty"); err == nil {
		defer f.Close()
		if w, _, err := term.GetSize(int(f.Fd())); err == nil && w > 0 {
			return w
		}
	}
	return 0
}
