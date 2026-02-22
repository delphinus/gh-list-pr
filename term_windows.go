//go:build windows

package main

import (
	"os"

	"golang.org/x/term"
)

func termWidthFromTTY() int {
	for _, f := range []*os.File{os.Stdin, os.Stdout, os.Stderr} {
		if w, _, err := term.GetSize(int(f.Fd())); err == nil && w > 0 {
			return w
		}
	}
	return 0
}
