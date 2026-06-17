package main

import "fmt"

// CheckStatus describes the outcome of comparing one desired output
// against the observed xrandr state.
type CheckStatus string

const (
    StatusOK            CheckStatus = "ok"
    StatusModeMismatch  CheckStatus = "mode_mismatch"
    StatusDisconnected  CheckStatus = "disconnected"
    StatusMissing       CheckStatus = "missing"        // in config, not in xrandr at all
    StatusUnexpectedOff CheckStatus = "unexpected_off" // connected but no active mode
    StatusShouldBeOff   CheckStatus = "should_be_off"  // config says off, but output is on
)

// CheckResult is the outcome of comparing one desired output against the
// observed xrandr state.
type CheckResult struct {
    Output   string
    Status   CheckStatus
    WantMode string
    GotMode  string
}

// Report aggregates all per-output results plus a top-level verdict.
// Messages is a human-readable summary kept for the current CLI; Results
// is the structured form used by the daemon and by structured logging.
type Report struct {
    Results  []CheckResult
    Messages []string
    OK       bool
}

// Compare returns a Report describing how the observed xrandr state compares
// to the desired configuration. It performs no I/O or logging.
func Compare(cfg Config, actual ActualState) Report {
    results := make([]CheckResult, 0, len(cfg.Layout.Outputs))
    messages := []string{}
    allOK := true

    for _, want := range cfg.Layout.Outputs {
        r := CheckResult{Output: want.Name, WantMode: want.Mode}

        got, found := actual.Outputs[want.Name]

        switch {
        case !found:
            r.Status = StatusMissing
        case want.Off:
            if got.Connected && got.Mode != "" {
                r.Status = StatusShouldBeOff
                r.GotMode = got.Mode
            } else {
                r.Status = StatusOK
            }
        case !got.Connected:
            r.Status = StatusDisconnected
        case got.Mode == "":
            r.Status = StatusUnexpectedOff
        case got.Mode != want.Mode:
            r.Status = StatusModeMismatch
            r.GotMode = got.Mode
        default:
            r.Status = StatusOK
            r.GotMode = got.Mode
        }

        if r.Status != StatusOK {
            allOK = false
            messages = append(messages, formatMessage(r))
        }
        results = append(results, r)
    }

    return Report{Results: results, Messages: messages, OK: allOK}
}

func formatMessage(r CheckResult) string {
    switch r.Status {
    case StatusModeMismatch:
        return fmt.Sprintf("%s: expected mode %s, got %s", r.Output, r.WantMode, r.GotMode)
    case StatusDisconnected:
        return fmt.Sprintf("%s: disconnected", r.Output)
    case StatusMissing:
        return fmt.Sprintf("%s: not present in xrandr output", r.Output)
    case StatusUnexpectedOff:
        return fmt.Sprintf("%s: connected but no active mode", r.Output)
    case StatusShouldBeOff:
        return fmt.Sprintf("%s: should be off but is showing mode %s", r.Output, r.GotMode)
    default:
        return fmt.Sprintf("%s: %s", r.Output, r.Status)
    }
}
