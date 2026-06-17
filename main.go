package main

import (
	"fmt"
	"log"
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
		log.Fatalf("load config: %v", err)
	}

	if err := ValidateConfig(cfg); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	out, err := exec.Command("xrandr", "--current").Output()
	if err != nil {
		log.Fatalf("run xrandr --current: %v", err)
	}

	actual, err := ParseXrandrCurrent(string(out))
	if err != nil {
		log.Fatalf("parse xrandr: %v", err)
	}

	result := Compare(cfg, actual)

	if result.OK {
		fmt.Println("ok")
		os.Exit(0)
	}

	fmt.Println("doesn't match desired state")

	for _, msg := range result.Messages {
		fmt.Println(" - " + msg)
	}

	os.Exit(1)
}

