package main

import (
	"bufio"
	"fmt"
	"strings"
)

type ActualOutput struct {
	Name      string
	Connected bool
	Mode      string

	// For later logging. Not filled yet by --current parser.
	EDID string
}

type ActualState struct {
	Outputs map[string]ActualOutput
}

func ParseXrandrCurrent(text string) (ActualState, error) {
	state := ActualState{
		Outputs: make(map[string]ActualOutput),
	}

	var currentName string

	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "" {
			continue
		}

		if isOutputHeader(line) {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return state, fmt.Errorf("invalid output header: %q", line)
			}

			name := fields[0]
			status := fields[1]

			out := ActualOutput{
				Name:      name,
				Connected: status == "connected",
			}

			state.Outputs[name] = out
			currentName = name

			continue
		}

		if currentName == "" {
			continue
		}

		// Mode lines are indented, for example:
		//    1920x1080     60.00*+  59.94
		//
		// The currently active mode has a '*' on one of its refresh-rate fields.
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}

			modeName := fields[0]

			for _, field := range fields[1:] {
				if strings.Contains(field, "*") {
					out := state.Outputs[currentName]
					out.Mode = modeName
					state.Outputs[currentName] = out
					break
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return state, err
	}

	return state, nil
}

func isOutputHeader(line string) bool {
	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return false
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return false
	}

	return fields[1] == "connected" || fields[1] == "disconnected"
}

