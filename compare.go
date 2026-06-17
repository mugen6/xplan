package main

import "fmt"

type CompareResult struct {
	OK       bool
	Messages []string
}

func Compare(cfg Config, actual ActualState) CompareResult {
	result := CompareResult{
		OK: true,
	}

	for _, desired := range cfg.Layout.Outputs {
		got, exists := actual.Outputs[desired.Name]

		if desired.Off {
			if !exists {
				// Desired off, and xrandr does not report it. Fine for now.
				continue
			}

			if got.Connected {
				result.OK = false
				result.Messages = append(
					result.Messages,
					fmt.Sprintf("%s: expected off/disconnected, got connected", desired.Name),
				)
			}

			continue
		}

		if !exists {
			result.OK = false
			result.Messages = append(
				result.Messages,
				fmt.Sprintf("%s: not present in xrandr output", desired.Name),
			)
			continue
		}

		if !got.Connected {
			result.OK = false
			result.Messages = append(
				result.Messages,
				fmt.Sprintf("%s: expected connected, got disconnected", desired.Name),
			)
			continue
		}

		if got.Mode == "" {
			result.OK = false
			result.Messages = append(
				result.Messages,
				fmt.Sprintf("%s: connected but no active mode found", desired.Name),
			)
			continue
		}

		if got.Mode != desired.Mode {
			result.OK = false
			result.Messages = append(
				result.Messages,
				fmt.Sprintf("%s: expected mode %s, got %s", desired.Name, desired.Mode, got.Mode),
			)
		}
	}

	return result
}

