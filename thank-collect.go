package main

import (
	"fmt"
	"os"
	"strings"

	"language-analysis/config"
	thanks "language-analysis/thank-analysis-src"
)

func main() {
	run := false
	for _, arg := range os.Args[1:] {
		if len(arg) > 0 && arg[0] == '-' {
			if eq := strings.Index(arg, "="); eq > 0 {
				config.Options[arg[1:eq]] = arg[eq+1:]
			} else {
				config.Options[arg[1:eq]] = ""
			}
			continue
		}
		switch arg {
		case "status":
			thanks.StatusCommand()
			run = true
		default:
			fmt.Fprintf(os.Stderr, "%s: unknown command %s.  Available commands: status\n", os.Args[0])
			os.Exit(1)
		}
	}
	if !run {
		count := 50
		if _, err := fmt.Sscanf(config.Options["thank-collect-count"], "%d", &count); err != nil {
			fmt.Printf("Error: %v", err)
			return
		}

		for range count {
			if !thanks.Collect() {
				break
			}
		}
	}
}
