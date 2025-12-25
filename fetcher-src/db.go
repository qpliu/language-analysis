package fetcher

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type fetcherDB struct {
	db *sql.DB
}

func openFetcherDB(dir string) (*fetcherDB, error) {
	db, err := sql.Open("sqlite3", dir+"/fetcher.db")
	if err != nil {
		return nil, err
	}

	fdb := fetcherDB{db}
	if err := fdb.init(); err != nil {
		fdb.Close()
		return nil, err
	}
	return &fdb, nil
}

func (db *fetcherDB) Close() error {
	return db.db.Close()
}

func (db *fetcherDB) init() error {
	if _, err := db.db.Query("SELECT feedID FROM feeds LIMIT 1"); err == nil {
		return nil
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, statement := range []string{
		`CREATE TABLE feeds (
			feedID INTEGER PRIMARY KEY AUTOINCREMENT,
			urlTemplate TEXT UNIQUE,
			scraperRx TEXT,
			scraperRxGroup INTEGER,
			earliestDateLimit DATE,

			earliestFetchDate DATE,
			earliestFetchDateTimestamp TIMESTAMP,
			latestFetchDate DATE,
			latestFetchDateTimestamp TIMESTAMP)`,
		`CREATE TABLE files (
			fileID INTEGER PRIMARY KEY AUTOINCREMENT,
			feedID INTEGER REFERENCES feed (feedID),
			url TEXT UNIQUE,
			date DATE,
			fetchTimestamp TIMESTAMP,
			purgeTimestamp TIMESTAMP)`,
		`CREATE INDEX filesFetchTimestamp ON files (fetchTimestamp)`,
		`CREATE INDEX filesPurgeTimestamp ON files (purgeTimestamp)`,
		`CREATE INDEX filesFeedIDFetchTimestamp ON files (feedID, fetchTimestamp)`,
		`CREATE INDEX filesFeedIDPurgeTimestamp ON files (feedID, purgeTimestamp)`,
		`CREATE INDEX filesDate ON files (date)`,
		`CREATE INDEX filesDatePurgeTimestamp ON files (date, purgeTimestamp)`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func parseDate(dateString sql.NullString) time.Time {
	t, _ := time.Parse(time.RFC3339, dateString.String)
	return t
}

func parseTimestamp(timestampString sql.NullString) time.Time {
	t, _ := time.Parse(time.RFC3339, timestampString.String)
	return t
}

func (db *fetcherDB) feeds() ([]Feed, error) {
	rows, err := db.db.Query("SELECT feedID, urlTemplate, scraperRx, scraperRxGroup, earliestDateLimit, earliestFetchDate, earliestFetchDateTimestamp, latestFetchDate, latestFetchDateTimestamp FROM feeds")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feeds := []Feed{}
	for rows.Next() {
		feed := Feed{}
		var earliestDateLimit sql.NullString
		var earliestFetchDate, earliestFetchDateTimestamp sql.NullString
		var latestFetchDate, latestFetchDateTimestamp sql.NullString
		if err := rows.Scan(&feed.feedID, &feed.urlTemplate, &feed.scraperRx, &feed.scraperRxGroup, &earliestDateLimit, &earliestFetchDate, &earliestFetchDateTimestamp, &latestFetchDate, &latestFetchDateTimestamp); err != nil {
			return nil, err
		}
		feed.earliestDateLimit = parseDate(earliestDateLimit)
		feed.earliestFetchDate = parseDate(earliestFetchDate)
		feed.earliestFetchDateTimestamp = parseTimestamp(earliestFetchDateTimestamp)
		feed.latestFetchDate = parseDate(latestFetchDate)
		feed.latestFetchDateTimestamp = parseTimestamp(latestFetchDateTimestamp)
		feeds = append(feeds, feed)
	}
	return feeds, nil
}

func (db *fetcherDB) feed(feedID int64) (Feed, error) {
	rows, err := db.db.Query("SELECT feedID, urlTemplate, scraperRx, scraperRxGroup, earliestDateLimit, earliestFetchDate, earliestFetchDateTimestamp, latestFetchDate, latestFetchDateTimestamp FROM feeds")
	if err != nil {
		return Feed{}, err
	}
	defer rows.Close()

	for rows.Next() {
		feed := Feed{}
		var earliestDateLimit sql.NullString
		var earliestFetchDate, earliestFetchDateTimestamp sql.NullString
		var latestFetchDate, latestFetchDateTimestamp sql.NullString
		if err := rows.Scan(&feed.feedID, &feed.urlTemplate, &feed.scraperRx, &feed.scraperRxGroup, &earliestDateLimit, &earliestFetchDate, &earliestFetchDateTimestamp, &latestFetchDate, &latestFetchDateTimestamp); err != nil {
			return Feed{}, err
		}
		feed.earliestDateLimit = parseDate(earliestDateLimit)
		feed.earliestFetchDate = parseDate(earliestFetchDate)
		feed.earliestFetchDateTimestamp = parseTimestamp(earliestFetchDateTimestamp)
		feed.latestFetchDate = parseDate(latestFetchDate)
		feed.latestFetchDateTimestamp = parseTimestamp(latestFetchDateTimestamp)
		return feed, nil
	}
	return Feed{}, fmt.Errorf("Nonexistent feedID: %d", feedID)
}

func (db *fetcherDB) file(fileID int64) (File, error) {
	rows, err := db.db.Query("SELECT fileID, feedID, url, date, fetchTimestamp, purgeTimestamp FROM files WHERE fileID = ?", fileID)
	if err != nil {
		return File{}, err
	}
	defer rows.Close()

	for rows.Next() {
		file := File{}
		var date sql.NullString
		var fetchTimestamp sql.NullString
		var purgeTimestamp sql.NullString
		if err := rows.Scan(&file.fileID, &file.feedID, &file.url, &date, &fetchTimestamp, &purgeTimestamp); err != nil {
			return File{}, err
		}
		file.date = parseDate(date)
		file.fetchTimestamp = parseTimestamp(fetchTimestamp)
		file.purgeTimestamp = parseTimestamp(purgeTimestamp)
		return file, nil
	}
	return File{}, fmt.Errorf("Nonexistent fileID: %d", fileID)
}

func (db *fetcherDB) unfetched(limit int) ([]File, error) {
	rows, err := db.db.Query("SELECT fileID, feedID, url, date, purgeTimestamp FROM files WHERE fetchTimestamp IS NULL LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []File{}
	for rows.Next() {
		file := File{}
		var date sql.NullString
		var purgeTimestamp sql.NullString
		if err := rows.Scan(&file.fileID, &file.feedID, &file.url, &date, &purgeTimestamp); err != nil {
			return nil, err
		}
		file.date = parseDate(date)
		file.purgeTimestamp = parseTimestamp(purgeTimestamp)
		files = append(files, file)
	}
	return files, nil
}

func (db *fetcherDB) fetchedSince(since time.Time, limit int) ([]File, error) {
	rows, err := db.db.Query("SELECT fileID, feedID, url, date, fetchTimestamp FROM files WHERE fetchTimestamp > ? AND purgeTimestamp IS NULL ORDER BY fetchTimestamp ASC LIMIT ?", since.Format(time.DateTime), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []File{}
	for rows.Next() {
		file := File{}
		var date sql.NullString
		var fetchTimestamp sql.NullString
		if err := rows.Scan(&file.fileID, &file.feedID, &file.url, &date, &fetchTimestamp); err != nil {
			return nil, err
		}
		file.date = parseDate(date)
		file.fetchTimestamp = parseTimestamp(fetchTimestamp)
		files = append(files, file)
	}
	return files, nil
}

func (db *fetcherDB) addFeed(urlTemplate, scraperRx string, scraperRxGroup int, earliestDateLimit time.Time) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT INTO feeds (urlTemplate, scraperRx, scraperRxGroup, earliestDateLimit) VALUES (?,?,?,?)", urlTemplate, scraperRx, scraperRxGroup, earliestDateLimit.Format(time.DateOnly)); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *fetcherDB) updateFeedEarliestFetched(feedID int64, earliestFetchedDate time.Time) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE feeds SET earliestFetchDate = ?, earliestFetchDateTimestamp = DATETIME() WHERE feedID = ?", earliestFetchedDate.Format(time.DateOnly), feedID); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *fetcherDB) updateFeedLatestFetched(feedID int64, latestFetchedDate time.Time) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE feeds SET latestFetchDate = ?, latestFetchDateTimestamp = DATETIME() WHERE feedID = ?", latestFetchedDate.Format(time.DateOnly), feedID); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *fetcherDB) addFile(feedID int64, url string, date time.Time) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT INTO files (feedID, url, date) VALUES (?,?,?)", feedID, url, date.Format(time.DateOnly)); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *fetcherDB) updateFileFetched(fileID int64) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE files SET fetchTimestamp = DATETIME(), purgeTimestamp = NULL WHERE fileID = ?", fileID); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *fetcherDB) earliestFetched(beforeTimestamp time.Time, limit int) ([]File, error) {
	rows, err := db.db.Query("SELECT fileID, feedID, url, date, fetchTimestamp FROM files WHERE fetchTimestamp < ? AND purgeTimestamp IS NULL ORDER BY fetchTimestamp ASC LIMIT ?", beforeTimestamp.Format(time.DateTime), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []File{}
	for rows.Next() {
		file := File{}
		var date sql.NullString
		var fetchTimestamp sql.NullString
		if err := rows.Scan(&file.fileID, &file.feedID, &file.url, &date, &fetchTimestamp); err != nil {
			return nil, err
		}
		file.date = parseDate(date)
		file.fetchTimestamp = parseTimestamp(fetchTimestamp)
		files = append(files, file)
	}
	return files, nil
}

func (db *fetcherDB) purgeFile(fileID int64) (bool, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	result, err := tx.Exec("UPDATE files SET purgeTimestamp = DATETIME() WHERE fileID = ?", fileID)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	if n, err := result.RowsAffected(); err != nil {
		return true, nil
	} else {
		return n == 1, nil
	}
}

func (db *fetcherDB) purged(minFileID int64, limit int) ([]File, error) {
	rows, err := db.db.Query("SELECT fileID, feedID, url, date, fetchTimestamp, purgeTimestamp FROM files WHERE fileID >= ? AND purgeTimestamp IS NOT NULL ORDER BY fileID ASC LIMIT ?", minFileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []File{}
	for rows.Next() {
		file := File{}
		var date sql.NullString
		var fetchTimestamp sql.NullString
		var purgeTimestamp sql.NullString
		if err := rows.Scan(&file.fileID, &file.feedID, &file.url, &date, &fetchTimestamp, &purgeTimestamp); err != nil {
			return nil, err
		}
		file.date = parseDate(date)
		file.fetchTimestamp = parseTimestamp(fetchTimestamp)
		file.purgeTimestamp = parseTimestamp(purgeTimestamp)
		files = append(files, file)
	}
	return files, nil
}

func (db *fetcherDB) reenqueue(fileID int64) (bool, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	result, err := tx.Exec("UPDATE files SET fetchTimestamp = NULL WHERE fileID = ?", fileID)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	if n, err := result.RowsAffected(); err != nil {
		return true, nil
	} else {
		return n == 1, nil
	}
}

func (db *fetcherDB) reenqueueDateRange(start, end time.Time) (int64, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec("UPDATE files SET fetchTimestamp = NULL WHERE date >= ? AND date <= ? AND purgeTimestamp IS NOT NULL", start.Format(time.DateOnly), end.Format(time.DateOnly))
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (db *fetcherDB) countUnfetched(feedID int64) (int, error) {
	rows, err := db.db.Query("SELECT COUNT(*) FROM files WHERE feedID = ? AND fetchTimestamp IS NULL", feedID)
	if err != nil {
		return 0, err
	}

	for rows.Next() {
		count := 0
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
		return count, nil
	}
	return 0, nil
}
