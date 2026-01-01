package thankAnalysis

import (
	"fmt"
	"time"

	"language-analysis/config"
	fetcher "language-analysis/fetcher-src"
	scraper "language-analysis/scraper-src"
)

func StatusCommand() error {
	db, err := openThankDB()
	if err != nil {
		return err
	}
	defer db.Close()

	fetchTimestamp, err := db.lastFetchTimestamp()
	if err != nil {
		return err
	}

	fmt.Printf("Last fetch timestamp: %s\n", fetchTimestamp.Format(time.DateTime))
	return nil
}

func CollectCommand() error {
	count, err := config.Int("thank-collect-count", 80)
	if err != nil {
		return err
	}

	db, err := openThankDB()
	if err != nil {
		return err
	}
	defer db.Close()

	for range count {
		fetchTimestamp, err := db.lastFetchTimestamp()
		if err != nil {
			return err
		}

		files, err := fetcher.FilesSince(fetchTimestamp, 1)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			fmt.Printf("No more files.\n")
			return nil
		}

		for _, file := range files {
			content, err := scraper.Scrape(file)
			if err != nil {
				return err
			}

			responses := map[string]map[[5]string]bool{}
			for _, resp := range ThankResponses(content) {
				fmt.Printf("%s,%d.%d: %s\n", file.Date().Format(time.DateOnly), file.ID(), resp.Index, resp)
				responses[resp.Name] = ResponsePhrases(resp.Text)
			}

			if err := db.addResponses(file.ID(), file.Date(), responses); err != nil {
				return err
			}

			if err := db.setFetchTimestamp(file.FetchTimestamp()); err != nil {
				return err
			}
		}
	}
	return nil
}
