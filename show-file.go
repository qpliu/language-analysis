package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"language-analysis/config"
	fetcher "language-analysis/fetcher-src"
	scraper "language-analysis/scraper-src"
)

func main() {
	for _, arg := range os.Args[1:] {
		if len(arg) > 0 && arg[0] == '-' {
			if eq := strings.Index(arg, "="); eq > 0 {
				config.Options[arg[1:eq]] = arg[eq+1:]
			} else {
				config.Options[arg[1:eq]] = ""
			}
			continue
		}
		var fileID int64
		if _, err := fmt.Sscanf(arg, "%d", &fileID); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else if file, err := fetcher.FileByID(fileID); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else if content, err := scraper.Scrape(file); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("%s %d\n", file.Date().Format(time.DateOnly), file.ID())
			for _, line := range content {
				fmt.Printf("  %s\n", line)
			}
		}
	}
}
