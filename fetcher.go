package main

import (
	"fmt"
	"os"
	"strings"

	"language-analysis/config"
	fetcher "language-analysis/fetcher-src"
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
			fetcher.StatusCommand()
			run = true
		case "fetch":
			fetcher.FetchCommand()
			run = true
		default:
			fmt.Fprintf(os.Stderr, "%s: unknown command %s.  Available commands: fetch status\n", os.Args[0])
			os.Exit(1)
		}
	}
	if !run {
		fetcher.FetchLoopCommand()
	}
}
