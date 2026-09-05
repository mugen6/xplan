package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// options holds parsed CLI flags. Passed around instead of using globals.
type options struct {
	ConfigPath  string
	ForceLogfmt bool
}

func main() {
	opts := parseFlags()

	// Build the logger once. useLogfmt decides whether we'll actually use it
	// for human-facing output, but fail() always uses it for error reporting
	// when we're not on a TTY.
	logger := setupLogger()
	slog.SetDefault(logger)

	cfg, err := LoadConfig(opts.ConfigPath)
	if err != nil {
		fail(opts, "load config failed", "path", opts.ConfigPath, "error", err)
	}

	out, err := exec.Command("xrandr", "--current").Output()
	if err != nil {
		fail(opts, "xrandr failed", "error", err)
	}

	actual, err := ParseXrandrCurrent(string(out))
	if err != nil {
		fail(opts, "parse xrandr failed", "error", err)
	}

	report := Compare(cfg, actual)
	reportToOutput(report, opts)

	if !report.OK {
		os.Exit(1)
	}
}

// parseFlags reads CLI arguments and returns an options struct.
func parseFlags() options {
	var opts options
	flag.StringVar(&opts.ConfigPath, "config", "monitors.kdl", "path to KDL config file")
	flag.BoolVar(&opts.ForceLogfmt, "logfmt", false, "force logfmt output even on a TTY (useful for testing)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags]\n\nflags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	return opts
}

// fail prints an error and exits with code 2 (operational error).
func fail(opts options, msg string, attrs ...any) {
	if opts.ForceLogfmt {
		slog.Error(msg, attrs...)
	} else {
		fmt.Fprintf(os.Stderr, "error: %s", msg)
		for i := 0; i+1 < len(attrs); i += 2 {
			fmt.Fprintf(os.Stderr, " %v=%v", attrs[i], attrs[i+1])
		}
		fmt.Fprintln(os.Stderr)
	}
	os.Exit(2)
}

// reportToOutput emits the comparison results: human-readable by default,
// logfmt if --logfmt was passed.
func reportToOutput(report Report, opts options) {
	if opts.ForceLogfmt {
		if report.OK {
			slog.Info("display state ok",
				"trigger", "oneshot",
				"outputs_checked", len(report.Results),
			)
			return
		}

		for _, r := range report.Results {
			if r.Status == StatusOK {
				continue
			}
			slog.Error("display output check failed",
				"trigger", "oneshot",
				"output", r.Output,
				"status", string(r.Status),
				"want_mode", r.WantMode,
				"got_mode", r.GotMode,
			)
		}
		return
	}

	for _, r := range report.Results {
		printResultLine(r)
	}
}
