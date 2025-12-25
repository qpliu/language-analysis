package phraseAnalysis

import (
	"fmt"
	"time"

	"language-analysis/config"
	fetcher "language-analysis/fetcher-src"
	scraper "language-analysis/scraper-src"
)

func StatusCommand() {
	db, err := openPhraseDB(config.Options["dir"])
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

func Collect(count int) bool {
	db, err := openPhraseDB(config.Options["dir"])
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

	files, err := fetcher.FilesSince(fetchTimestamp, count)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}
	if len(files) == 0 {
		fmt.Printf("No more files.\n")
		return false
	}

	phraseTotals := 0
	prefaceTotals := 0
	for _, file := range files {
		content, err := scraper.Scrape(file)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}

		phraseCounts, prefaceCounts := CountPhrases(content)
		if err := db.addCounts(file.ID(), file.Date(), phraseCounts, prefaceCounts); err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}

		if err := db.setFetchTimestamp(file.FetchTimestamp()); err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}

		for _, count := range phraseCounts {
			phraseTotals += count
		}
		for _, count := range prefaceCounts {
			prefaceTotals += count
		}
	}
	fmt.Printf("Counted %d phrase(s), %d preface(s).\n", phraseTotals, prefaceTotals)
	return true
}
