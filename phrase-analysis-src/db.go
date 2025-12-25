package phraseAnalysis

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type phraseDB struct {
	db *sql.DB
}

func openPhraseDB(dir string) (*phraseDB, error) {
	db, err := sql.Open("sqlite3", dir+"/phrase-analysis.db")
	if err != nil {
		return nil, err
	}

	tdb := phraseDB{db}
	if err := tdb.init(); err != nil {
		tdb.Close()
		return nil, err
	}
	return &tdb, nil
}

func (db *phraseDB) Close() error {
	return db.db.Close()
}

func (db *phraseDB) init() error {
	if _, err := db.db.Query("SELECT lastFetchTimestamp FROM fetcherState LIMIT 1"); err == nil {
		return nil
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, statement := range []string{
		`CREATE TABLE fetcherState (
			lastFetchTimestamp TIMESTAMP)`,
		`INSERT INTO fetcherState (lastFetchTimestamp)
			VALUES ('1970-01-01 00:00:00')`,
		`CREATE TABLE phrases (
			phraseID INTEGER PRIMARY KEY,
			phrase TEXT UNIQUE)`,
		`CREATE TABLE prefaces (
			prefaceID INTEGER PRIMARY KEY,
			preface TEXT UNIQUE)`,
		`CREATE TABLE speakers (
			speakerID INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE)`,
		`CREATE INDEX speakersName ON speakers (name)`,
		`CREATE TABLE files (
			fileID INTEGER PRIMARY KEY,
			date DATE)`,
		`CREATE INDEX fileDate ON files (date)`,
		`CREATE TABLE phraseCounts (
			fileID INTEGER REFERENCES files (fileID),
			speakerID INTEGER REFERENCES speakers (speakerID),
			phraseID INTEGER REFERENCES phrases (phraseID),
			count INTEGER)`,
		`CREATE INDEX phraseCountsSpeakerID ON phraseCounts (speakerID)`,
		`CREATE INDEX phraseCountsPhraseID ON phraseCounts (phraseID)`,
		`CREATE TABLE prefaceCounts (
			fileID INTEGER REFERENCES files (fileID),
			speakerID INTEGER REFERENCES speakers (speakerID),
			prefaceID INTEGER REFERENCES prefaces (prefaceID),
			count INTEGER)`,
		`CREATE INDEX prefaceCountsSpeakerID ON prefaceCounts (speakerID)`,
		`CREATE INDEX prefaceCountsPrefaceID ON prefaceCounts (prefaceID)`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}

	for i, phrase := range PHRASES {
		if _, err := tx.Exec("INSERT INTO phrases (phraseID, phrase) VALUES (?,?)", i, phrase); err != nil {
			return err
		}
	}

	for i, preface := range PREFACES {
		if _, err := tx.Exec("INSERT INTO prefaces (prefaceID, preface) VALUES (?,?)", i, preface); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func parseDate(dateString sql.NullString) time.Time {
	t, _ := time.Parse(time.RFC3339, dateString.String)
	return t
}

func parseTimestamp(timestampString sql.NullString) time.Time {
	t, _ := time.Parse(time.RFC3339, timestampString.String)
	return t
}

func (db *phraseDB) lastFetchTimestamp() (time.Time, error) {
	rows, err := db.db.Query("SELECT lastFetchTimestamp FROM fetcherState LIMIT 1")
	if err != nil {
		return time.Time{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var fetchTimestamp sql.NullString
		if err := rows.Scan(&fetchTimestamp); err != nil {
			return time.Time{}, err
		}
		return parseTimestamp(fetchTimestamp), nil
	}

	tx, err := db.db.Begin()
	if err != nil {
		return time.Time{}, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT INTO fetcherState (lastFetchTimestamp) VALUES ('1970-01-01 00:00:00')"); err != nil {
		return time.Time{}, err
	}

	return time.Time{}, tx.Commit()
}

func (db *phraseDB) setFetchTimestamp(lastFetchTimestamp time.Time) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE fetcherState SET lastFetchTimestamp = ?", lastFetchTimestamp.Format(time.DateTime)); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *phraseDB) addCounts(fileID int64, date time.Time, phrases map[[2]string]int, prefaces map[[2]string]int) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT INTO files (fileID, date) VALUES (?,?)", fileID, date.Format(time.DateOnly)); err != nil {
		return err
	}

	speakerIDs := map[string]int64{}
	for item := range phrases {
		if _, ok := speakerIDs[item[0]]; ok {
			continue
		}
		speakerID, err := db.getSpeakerID(tx, item[0])
		if err != nil {
			return err
		}
		speakerIDs[item[0]] = speakerID
	}
	for item := range prefaces {
		if _, ok := speakerIDs[item[0]]; ok {
			continue
		}
		speakerID, err := db.getSpeakerID(tx, item[0])
		if err != nil {
			return err
		}
		speakerIDs[item[0]] = speakerID
	}

	for item, count := range phrases {
		phraseID, ok := PHRASEIDS[item[1]]
		if !ok || count <= 0 {
			continue
		}
		if _, err := tx.Exec("INSERT INTO phraseCounts (fileID, speakerID, phraseID, count) VALUES (?,?,?,?)", fileID, speakerIDs[item[0]], phraseID, count); err != nil {
			return nil
		}
	}

	for item, count := range prefaces {
		prefaceID, ok := PREFACEIDS[item[1]]
		if !ok || count <= 0 {
			continue
		}
		if _, err := tx.Exec("INSERT INTO prefaceCounts (fileID, speakerID, prefaceID, count) VALUES (?,?,?,?)", fileID, speakerIDs[item[0]], prefaceID, count); err != nil {
			return nil
		}
	}

	return tx.Commit()
}

func (db *phraseDB) getSpeakerID(tx *sql.Tx, speaker string) (int64, error) {
	for range 2 {
		rows, err := tx.Query("SELECT speakerID FROM speakers WHERE name = ?", speaker)
		if err != nil {
			return 0, err
		}

		for rows.Next() {
			var speakerID int64
			if err := rows.Scan(&speakerID); err != nil {
				return 0, err
			}
			return speakerID, nil
		}

		result, err := tx.Exec("INSERT INTO speakers (name) VALUES (?)", speaker)
		if err != nil {
			return 0, err
		}

		if speakerID, err := result.LastInsertId(); err == nil {
			return speakerID, nil
		}
	}
	return 0, fmt.Errorf("Failed to get speakerID for %s", speaker)
}
