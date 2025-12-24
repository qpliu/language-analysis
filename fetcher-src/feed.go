package fetcher

import (
	"fmt"
	"regexp"
	"time"

	"language-analysis/config"
)

type Feed struct {
	feedID            int64
	urlTemplate       string
	scraperRx         string
	scraperRxGroup    int
	earliestDateLimit time.Time

	earliestFetchDate          time.Time
	earliestFetchDateTimestamp time.Time
	latestFetchDate            time.Time
	latestFetchDateTimestamp   time.Time
}

func AddFeed(urlTemplate, scraperRx string, scraperRxGroup int, earliestDateLimit time.Time) error {
	db, err := openFetcherDB(config.Options["dir"])
	if err != nil {
		return err
	}
	defer db.Close()

	return db.addFeed(urlTemplate, scraperRx, scraperRxGroup, earliestDateLimit)
}

func fetchFeed(feed Feed, db *fetcherDB) bool {
	feedRegex, err := regexp.Compile(feed.scraperRx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}

	earliest := true
	fetchDate := feed.earliestFetchDate

	if feed.earliestFetchDateTimestamp.Before(feed.latestFetchDateTimestamp) {
		fetchDate = feed.earliestFetchDate
		if fetchDate.IsZero() {
			fetchDate = feed.latestFetchDate
		}
	} else {
		earliest = false
		fetchDate = feed.latestFetchDate
		if fetchDate.IsZero() {
			fetchDate = feed.earliestFetchDate
		}
	}
	if fetchDate.IsZero() {
		fetchDate = time.Now().AddDate(0, 0, -7)
	} else if earliest {
		fetchDate = fetchDate.AddDate(0, 0, -1)
		if fetchDate.Before(feed.earliestDateLimit) {
			db.updateFeedEarliestFetched(feed.feedID, fetchDate.AddDate(0, 0, 1))
			return false
		}
	} else {
		fetchDate = fetchDate.AddDate(0, 0, 1)
		if fetchDate.After(time.Now().AddDate(0, 0, -7)) {
			db.updateFeedLatestFetched(feed.feedID, fetchDate.AddDate(0, 0, -1))
			return false
		}
	}

	feedData, err := fetch(fmt.Sprintf(feed.urlTemplate, fetchDate.Format(time.DateOnly)))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}

	count := 0
	for _, match := range feedRegex.FindAllSubmatchIndex(feedData, -1) {
		if len(match) > 2*feed.scraperRxGroup+1 {
			link := string(feedData[match[2*feed.scraperRxGroup]:match[2*feed.scraperRxGroup+1]])
			if err := db.addFile(feed.feedID, link, fetchDate); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				count++
			}
		}
	}
	fmt.Printf("enqueued %d file(s).\n", count)

	if earliest {
		err := db.updateFeedEarliestFetched(feed.feedID, fetchDate)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		db.updateFeedLatestFetched(feed.feedID, fetchDate)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
	return true
}
