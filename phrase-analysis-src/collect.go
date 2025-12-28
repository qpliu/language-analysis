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

	phraseTotals := 0
	prefaceTotals := 0
	for range count {
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
			fmt.Printf("Counted %d phrase(s), %d preface(s).\n", phraseTotals, prefaceTotals)
			fmt.Printf("No more files.\n")
			return false
		}

		content, err := scraper.Scrape(files[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}

		phrases, prefaces, err := db.unfetchedPhrasesPrefaces(files[0].FetchTimestamp())
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}

		phraseCounts, prefaceCounts := CountPhrases(content, phrases, prefaces)
		if err := db.addCounts(files[0].ID(), files[0].Date(), phrases, prefaces, phraseCounts, prefaceCounts); err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}

		if err := db.setFetchTimestamp(files[0].FetchTimestamp()); err != nil {
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
