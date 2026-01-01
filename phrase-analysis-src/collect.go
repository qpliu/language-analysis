package phraseAnalysis

import (
	"fmt"
	"time"

	"language-analysis/config"
	fetcher "language-analysis/fetcher-src"
	scraper "language-analysis/scraper-src"
)

func StatusCommand() error {
	db, err := openPhraseDB()
	if err != nil {
		return err
	}
	defer db.Close()

	fetchTimestamp, err := db.lastFetchTimestamp()
	if err != nil {
		return err
	}

	fmt.Printf("Last fetch timestamp: %s\n", fetchTimestamp.Format(time.DateTime))

	phraseStatus := map[string]int{}
	prefaceStatus := map[string]int{}
	for _, phrase := range Config.Phrases {
		phraseStatus[phrase] += 1
	}
	for _, preface := range Config.Prefaces {
		prefaceStatus[preface] += 1
	}

	dbPhrases, dbPrefaces, err := db.phrasesPrefaces()
	if err != nil {
		return err
	}
	for phrase := range dbPhrases {
		phraseStatus[phrase] += 2
	}
	for preface := range dbPrefaces {
		prefaceStatus[preface] += 2
	}

	unaddedPhrases, unconfiguredPhrases := 0, 0
	for _, status := range phraseStatus {
		switch status {
		case 1:
			unaddedPhrases++
		case 2:
			unconfiguredPhrases++
		}
	}
	if unaddedPhrases != 0 {
		fmt.Printf("%d unadded phrase(s), %d unconfigured database phrases(s).\n", unaddedPhrases, unconfiguredPhrases)
	}

	unaddedPrefaces, unconfiguredPrefaces := 0, 0
	for _, status := range prefaceStatus {
		switch status {
		case 1:
			unaddedPrefaces++
		case 2:
			unconfiguredPrefaces++
		}
	}
	if unaddedPrefaces != 0 {
		fmt.Printf("%d unadded preface(s), %d unconfigured database prefacse(s).\n", unaddedPrefaces, unconfiguredPrefaces)
	}
	return nil
}

func CollectCommand() error {
	count, err := config.Int("phrase-collect-count", 1000)
	if err != nil {
		return err
	}

	db, err := openPhraseDB()
	if err != nil {
		return err
	}
	defer db.Close()

	phraseTotals := 0
	prefaceTotals := 0
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
			fmt.Printf("Counted %d phrase(s), %d preface(s).\n", phraseTotals, prefaceTotals)
			fmt.Printf("No more files.\n")
			return nil
		}

		content, err := scraper.Scrape(files[0])
		if err != nil {
			return err
		}

		phrases, prefaces, err := db.unfetchedPhrasesPrefaces(files[0].FetchTimestamp())
		if err != nil {
			return err
		}

		phraseCounts, prefaceCounts := CountPhrases(content, phrases, prefaces)
		if err := db.addCounts(files[0].ID(), files[0].Date(), phrases, prefaces, phraseCounts, prefaceCounts); err != nil {
			return err
		}

		if err := db.setFetchTimestamp(files[0].FetchTimestamp()); err != nil {
			return err
		}

		for _, count := range phraseCounts {
			phraseTotals += count
		}
		for _, count := range prefaceCounts {
			prefaceTotals += count
		}
	}
	fmt.Printf("Counted %d phrase(s), %d preface(s).\n", phraseTotals, prefaceTotals)
	return nil
}

func AddCommand() error {
	db, err := openPhraseDB()
	if err != nil {
		return err
	}
	defer db.Close()

	dbPhrases, dbPrefaces, err := db.phrasesPrefaces()
	if err != nil {
		return err
	}

	phrasesAdded := 0
	for _, phrase := range Config.Phrases {
		if _, ok := dbPhrases[phrase]; ok {
			continue
		}
		if err := AddPhrase(phrase); err != nil {
			return err
		}
		phrasesAdded++
	}

	prefacesAdded := 0
	for _, preface := range Config.Prefaces {
		if _, ok := dbPrefaces[preface]; ok {
			continue
		}
		if err := AddPreface(preface); err != nil {
			return err
		}
		prefacesAdded++
	}

	fmt.Printf("Added %d phrase(s), %d preface(s).\n", phrasesAdded, prefacesAdded)
	return nil
}
