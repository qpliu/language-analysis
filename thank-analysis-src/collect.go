package thankAnalysis

import (
	"fmt"
	"time"

	"language-analysis/config"
	fetcher "language-analysis/fetcher-src"
	scraper "language-analysis/scraper-src"
)

func StatusCommand() {
	db, err := openThankDB(config.Options["dir"])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer db.Close()

	fetchTimestamp, err := db.lastFetchTimestamp()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Last fetch timestamp: %s\n", fetchTimestamp.Format(time.DateTime))
}

func Collect() bool {
	db, err := openThankDB(config.Options["dir"])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}
	defer db.Close()

	fetchTimestamp, err := db.lastFetchTimestamp()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}

	files, err := fetcher.FilesSince(fetchTimestamp, 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}
	if len(files) == 0 {
		fmt.Printf("No more files.\n")
		return false
	}

	for _, file := range files {
		content, err := scraper.Scrape(file)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}

		responses := map[string]map[[5]string]bool{}
		for _, resp := range ThankResponses(content) {
			fmt.Printf("%s,%d.%d: %s\n", file.Date().Format(time.DateOnly), file.ID(), resp.Index, resp)
			responses[resp.Name] = ResponsePhrases(resp.Text)
		}

		if err := db.addResponses(file.ID(), file.Date(), responses); err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}

		if err := db.setFetchTimestamp(file.FetchTimestamp()); err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}
	}
	return true
}
