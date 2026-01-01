package fetcher

import (
	"fmt"
	"time"

	"language-analysis/config"
)

var Config struct {
	Feed []struct {
		Name              string
		URLTemplate       string
		ScraperRx         string
		ScraperRxGroup    int
		EarliestDateLimit time.Time
	}
}

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

func StatusCommand() error {
	db, err := openFetcherDB()
	if err != nil {
		return fmt.Errorf("Failed to open fetcher database: %v", err)
	}
	defer db.Close()
	feeds, err := db.feeds()
	if err != nil {
		return fmt.Errorf("Failed to get feeds from fetcher database: %v", err)
	}

	unaddedFeeds := []string{}
	for _, f := range Config.Feed {
		unadded := true
		for _, feed := range feeds {
			if f.URLTemplate == feed.urlTemplate {
				unadded = false
				break
			}
		}
		if unadded {
			unaddedFeeds = append(unaddedFeeds, f.Name)
		}
	}
	if len(unaddedFeeds) > 0 {
		fmt.Printf("Unadded feeds:\n")
		for i, f := range unaddedFeeds {
			fmt.Printf(" %2d: %s\n", i, f)
		}
	}

	for _, feed := range feeds {
		unconfigured := true
		for _, f := range Config.Feed {
			if f.URLTemplate == feed.urlTemplate {
				unconfigured = false
				fmt.Printf("Feed %d: %s\n", feed.feedID, f.Name)
				break
			}
		}
		if unconfigured {
			fmt.Printf("Feed %d (unconfigured): %s\n", feed.feedID, feed.urlTemplate)
		}
		fmt.Printf("    earliest: %s (%s)\n", formatDate(feed.earliestFetchDate), formatTimestamp(feed.earliestFetchDateTimestamp))
		fmt.Printf("    latest: %s (%s)\n", formatDate(feed.latestFetchDate), formatTimestamp(feed.latestFetchDateTimestamp))
		if count, err := db.countUnfetched(feed.feedID); err != nil {
			fmt.Printf("    error fetching pending unfetched count: %v\n", err)
		} else if count > 0 {
			fmt.Printf("    pending unfetched count: %d\n", count)
		}
	}
	return nil
}

func FetchCommand() error {
	db, err := openFetcherDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return fetchNext(db)
}

func AddFeedsCommand() error {
	return fmt.Errorf("Not implemented.")
}

func FetchLoopCommand() error {
	sleep, err := config.Duration("fetcher-sleep", 15*time.Second)
	if err != nil {
		return err
	}

	db, err := openFetcherDB()
	if err != nil {
		return err
	}
	defer db.Close()

	for {
		if err := fetchNext(db); err != nil {
			return err
		}
		time.Sleep(sleep)
	}
}

func fetchNext(db *fetcherDB) error {
	files, err := db.unfetched(10)
	if err != nil {
		return err
	}

	if len(files) > 0 {
		var err error
		for _, file := range files {
			if err = fetchFile(file, db); err == nil {
				return nil
			}
		}
		return err
	}

	feeds, err := db.feeds()
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		return nil
	}

	feed := feeds[0]
	for _, f := range feeds[1:] {
		if feed.earliestFetchDateTimestamp.IsZero() {
			if fetched, err := fetchFeed(feed, db); err != nil {
				return err
			} else if fetched {
				return nil
			}
			feed = f
		} else if f.earliestFetchDateTimestamp.Before(feed.earliestFetchDateTimestamp) {
			feed = f
		}
	}
	_, err = fetchFeed(feed, db)
	return err
}
