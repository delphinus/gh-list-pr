package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
)

// 8-dot braille spinner, same as gh CLI (briandowns/spinner CharSets[11]).
var spinnerFrames = [...]string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

type spinner struct {
	msg  string
	quit chan struct{}
	wg   sync.WaitGroup
}

func newSpinner(msg string) *spinner {
	return &spinner{
		msg:  msg,
		quit: make(chan struct{}),
	}
}

func (s *spinner) start() {
	if !isatty.IsTerminal(os.Stderr.Fd()) && !isatty.IsCygwinTerminal(os.Stderr.Fd()) {
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		i := 0
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()

		// Draw the first frame immediately.
		fmt.Fprintf(os.Stderr, "\r%s %s", spinnerFrames[i], s.msg)
		i++

		for {
			select {
			case <-s.quit:
				// Clear the line.
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			case <-ticker.C:
				fmt.Fprintf(os.Stderr, "\r%s %s", spinnerFrames[i%len(spinnerFrames)], s.msg)
				i++
			}
		}
	}()
}

func (s *spinner) stop() {
	select {
	case <-s.quit:
		// Already stopped.
	default:
		close(s.quit)
	}
	s.wg.Wait()
}
