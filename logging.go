package main

import (
	"fmt"
	"log/slog"
	"os"

	"golang.org/x/term"
)

// isTTY reports whether stdout is connected to a terminal.
func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// setupLogger returns a slog.Logger emitting logfmt-style text to stdout,
// without a time field (the log collector adds its own).
func setupLogger() *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	return slog.New(handler)
}

// printResultLine writes a single human-readable line describing one
// CheckResult to stdout. Used for interactive (TTY) output.
func printResultLine(r CheckResult) {
	switch r.Status {
	case StatusOK:
		fmt.Printf("OK: %s %s\n", r.Output, r.GotMode)
	case StatusModeMismatch:
		fmt.Printf("FAIL: %s expected %s, got %s\n", r.Output, r.WantMode, r.GotMode)
	case StatusDisconnected:
		fmt.Printf("FAIL: %s disconnected\n", r.Output)
	case StatusMissing:
		fmt.Printf("FAIL: %s not present in xrandr output\n", r.Output)
	case StatusUnexpectedOff:
		fmt.Printf("FAIL: %s connected but no active mode\n", r.Output)
	case StatusShouldBeOff:
		fmt.Printf("FAIL: %s should be off but is showing %s\n", r.Output, r.GotMode)
	}
}
