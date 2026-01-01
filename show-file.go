package main

import (
	"fmt"
	"time"

	"language-analysis/config"
	fetcher "language-analysis/fetcher-src"
	scraper "language-analysis/scraper-src"
)

func main() {
	config.Run(nil, config.Command{
		Run: func() error {
			return nil
		},
	}, nil, showFile)
}

func showFile(arg string) (bool, error) {
	var fileID int64
	if _, err := fmt.Sscanf(arg, "%d", &fileID); err != nil {
		return false, err
	} else if file, err := fetcher.FileByID(fileID); err != nil {
		return false, err
	} else if content, err := scraper.Scrape(file); err != nil {
		return false, err
	} else {
		fmt.Printf("%s %d\n", file.Date().Format(time.DateOnly), file.ID())
		for _, line := range content {
			fmt.Printf("  %s\n", line)
		}
		return true, nil
	}
}
