package main

import (
	"fmt"
	"os"

	kdl "github.com/njreid/gokdl2"
)

type Config struct {
	Layout Layout `kdl:"layout"`
}

type Layout struct {
	Outputs []DesiredOutput `kdl:"output,multiple"`
}

type DesiredOutput struct {
	Name string `kdl:",arg"`

	// Basic fields used now.
	Mode string `kdl:"mode,omitempty"`
	Off  bool   `kdl:"off,omitempty"`

	// Future fields. Keeping these here is fine.
	Rate     float64   `kdl:"rate,omitempty"`
	Rotation string    `kdl:"rotation,omitempty"`
	Pos      []int     `kdl:"pos,omitempty"`
	Scale    []float64 `kdl:"scale,omitempty"`
	RightOf  string    `kdl:"right-of,omitempty"`
	LeftOf   string    `kdl:"left-of,omitempty"`
	Above    string    `kdl:"above,omitempty"`
	Below    string    `kdl:"below,omitempty"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config

	if err := kdl.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func ValidateConfig(cfg Config) error {
	seen := make(map[string]bool)

	if len(cfg.Layout.Outputs) == 0 {
		return fmt.Errorf("layout must contain at least one output")
	}

	for _, out := range cfg.Layout.Outputs {
		if out.Name == "" {
			return fmt.Errorf("output with empty name")
		}

		if seen[out.Name] {
			return fmt.Errorf("duplicate output %q", out.Name)
		}
		seen[out.Name] = true

		if out.Off {
			if out.Mode != "" {
				return fmt.Errorf("%s: off output must not have mode", out.Name)
			}
			continue
		}

		if out.Mode == "" {
			return fmt.Errorf("%s: enabled output must have mode", out.Name)
		}

		if len(out.Scale) != 0 && len(out.Scale) != 1 && len(out.Scale) != 2 {
			return fmt.Errorf("%s: scale must have one or two numbers", out.Name)
		}

		if len(out.Pos) != 0 && len(out.Pos) != 2 {
			return fmt.Errorf("%s: pos must have exactly two integers", out.Name)
		}

		relativeCount := countNonEmpty(out.LeftOf, out.RightOf, out.Above, out.Below)
		if relativeCount > 1 {
			return fmt.Errorf("%s: use only one of left-of, right-of, above, below", out.Name)
		}

		if len(out.Pos) != 0 && relativeCount != 0 {
			return fmt.Errorf("%s: use either pos or relative positioning, not both", out.Name)
		}

		switch out.Rotation {
		case "", "normal", "left", "right", "inverted":
		default:
			return fmt.Errorf("%s: invalid rotation %q", out.Name, out.Rotation)
		}
	}

	return nil
}

func countNonEmpty(values ...string) int {
	n := 0

	for _, value := range values {
		if value != "" {
			n++
		}
	}

	return n
}
