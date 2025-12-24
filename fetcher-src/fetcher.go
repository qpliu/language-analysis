package fetcher

import (
	"fmt"
	"time"

	"language-analysis/config"
)

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "NONE"
	}
	return t.Format(time.DateOnly)
}

func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "NONE"
	}
	return t.Format(time.DateTime)
}

func StatusCommand() {
	db, err := openFetcherDB(config.Options["dir"])
	if err != nil {
		fmt.Printf("Failed to open fetcher database: %v\n", err)
		return
	}
	feeds, err := db.feeds()
	if err != nil {
		fmt.Printf("Failed to get feeds from fetcher database: %v\n", err)
		return
	}
	for _, feed := range feeds {
		fmt.Printf("Feed %d:\n", feed.feedID)
		fmt.Printf("    url template: %s\n", feed.urlTemplate)
		fmt.Printf("    scraper rx(%d): %s\n", feed.scraperRxGroup, feed.scraperRx)
		fmt.Printf("    date limit: %s\n", formatDate(feed.earliestDateLimit))
		fmt.Printf("    earliest: %s (%s)\n", formatDate(feed.earliestFetchDate), formatTimestamp(feed.earliestFetchDateTimestamp))
		fmt.Printf("    latest: %s (%s)\n", formatDate(feed.latestFetchDate), formatTimestamp(feed.latestFetchDateTimestamp))
		if count, err := db.countUnfetched(feed.feedID); err != nil {
			fmt.Printf("    error fetching pending unfetched count: %v\n", err)
		} else {
			fmt.Printf("    pending unfetched count: %d\n", count)
		}
	}
}

func FetchCommand() {
	db, err := openFetcherDB(config.Options["dir"])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer db.Close()
	fetchNext(db)
}

func FetchLoopCommand() {
	sleep, err := time.ParseDuration(config.Options["fetcher-sleep"])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	db, err := openFetcherDB(config.Options["dir"])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer db.Close()

	for {
		fetchNext(db)
		time.Sleep(sleep)
	}
}

func fetchNext(db *fetcherDB) {
	files, err := db.unfetched(10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(files) > 0 {
		for _, file := range files {
			if fetchFile(file, db) {
				return
			}
		}
		return
	}

	feeds, err := db.feeds()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	if len(feeds) == 0 {
		return
	}

	feed := feeds[0]
	for _, f := range feeds[1:] {
		if feed.earliestFetchDateTimestamp.IsZero() {
			if fetchFeed(feed, db) {
				return
			}
			feed = f
		} else if f.earliestFetchDateTimestamp.Before(feed.earliestFetchDateTimestamp) {
			feed = f
		}
	}
	fetchFeed(feed, db)
}
