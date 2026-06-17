package main

import (
    // "log/slog"
    "os"
    "os/exec"
)

func main() {
    configPath := "monitors.kdl"
    if len(os.Args) > 1 {
        configPath = os.Args[1]
    }

    cfg, err := LoadConfig(configPath)
    if err != nil {
        fail("load config failed", "path", configPath, "error", err.Error())
    }
    if err := ValidateConfig(cfg); err != nil {
        fail("invalid config", "path", configPath, "error", err.Error())
    }

    out, err := exec.Command("xrandr", "--current").Output()
    if err != nil {
        fail("xrandr failed", "error", err.Error())
    }

    actual, err := ParseXrandrCurrent(string(out))
    if err != nil {
        fail("parse xrandr failed", "error", err.Error())
    }

    report := Compare(cfg, actual)
    reportToOutput(report)

    if report.OK {
        os.Exit(0)
    }
    os.Exit(1)
}

// reportToOutput writes the report to stdout in TTY-friendly text when
// interactive, or structured JSON otherwise.
func reportToOutput(report Report) {
    if isTTY() {
        for _, r := range report.Results {
            printResultLine(r)
        }
        return
    }

    logger := setupLogger()
    if report.OK {
        logger.Info("display state ok",
            "trigger", "oneshot",
            "outputs_checked", len(report.Results),
        )
        return
    }
    for _, r := range report.Results {
        if r.Status == StatusOK {
            continue
        }
        logger.Error("display output check failed",
            "trigger", "oneshot",
            "output", r.Output,
            "status", string(r.Status),
            "want_mode", r.WantMode,
            "got_mode", r.GotMode,
        )
    }
}

// fail emits an error in the appropriate format and exits with status 2,
// reserved for operational errors (config invalid, xrandr unavailable, etc.)
// as distinct from a degraded display state (exit 1).
func fail(msg string, kv ...string) {
    if isTTY() {
        os.Stderr.WriteString("error: " + msg)
        for i := 0; i+1 < len(kv); i += 2 {
            os.Stderr.WriteString(" " + kv[i] + "=" + kv[i+1])
        }
        os.Stderr.WriteString("\n")
    } else {
        logger := setupLogger()
        attrs := make([]any, 0, len(kv))
        for i := 0; i+1 < len(kv); i += 2 {
            attrs = append(attrs, kv[i], kv[i+1])
        }
        logger.Error(msg, attrs...)
    }
    os.Exit(2)
}

