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
	if _, err := db.db.Query("SELECT phraseID FROM phrases LIMIT 1"); err == nil {
		return nil
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, statement := range []string{
		`CREATE TABLE phrases (
			phraseID INTEGER PRIMARY KEY AUTOINCREMENT,
			phrase TEXT UNIQUE NOT NULL,
			lastFetchTimestamp TIMESTAMP)`,
		`CREATE INDEX phrasesPhrase ON phrases (phrase)`,
		`CREATE INDEX phrasesLastFetchTimestamp ON phrases (lastFetchTimestamp)`,
		`CREATE TABLE prefaces (
			prefaceID INTEGER PRIMARY KEY AUTOINCREMENT,
			preface TEXT UNIQUE NOT NULL,
			lastFetchTimestamp TIMESTAMP)`,
		`CREATE INDEX prefacesPreface ON prefaces (preface)`,
		`CREATE INDEX prefacesLastFetchTimestamp ON prefaces (lastFetchTimestamp)`,
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

	return tx.Commit()
}

func parseDate(dateString sql.NullString) time.Time {
	if t, err := time.Parse(time.RFC3339, dateString.String); err == nil {
		return t
	}
	t, _ := time.Parse(time.DateOnly, dateString.String)
	return t
}

func parseTimestamp(timestampString sql.NullString) time.Time {
	if t, err := time.Parse(time.RFC3339, timestampString.String); err == nil {
		return t
	}
	t, _ := time.Parse(time.DateTime, timestampString.String)
	return t
}

func (db *phraseDB) lastFetchTimestamp() (time.Time, error) {
	rows, err := db.db.Query("SELECT MIN(phrases.lastFetchTimestamp),  MIN(prefaces.lastFetchTimestamp) FROM phrases, prefaces")
	if err != nil {
		return time.Time{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var phrasesFetchTimestamp sql.NullString
		var prefacesFetchTimestamp sql.NullString
		if err := rows.Scan(&phrasesFetchTimestamp, &prefacesFetchTimestamp); err != nil {
			return time.Time{}, err
		}
		ts1 := parseTimestamp(phrasesFetchTimestamp)
		ts2 := parseTimestamp(prefacesFetchTimestamp)
		if ts1.Before(ts2) {
			return ts1, nil
		} else {
			return ts2, nil
		}
	}
	return time.Time{}, nil
}

func (db *phraseDB) setFetchTimestamp(lastFetchTimestamp time.Time) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE phrases SET lastFetchTimestamp = ? WHERE lastFetchTimestamp < ?", lastFetchTimestamp.Format(time.DateTime), lastFetchTimestamp.Format(time.DateTime)); err != nil {
		return err
	}

	if _, err := tx.Exec("UPDATE prefaces SET lastFetchTimestamp = ? WHERE lastFetchTimestamp < ?", lastFetchTimestamp.Format(time.DateTime), lastFetchTimestamp.Format(time.DateTime)); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *phraseDB) unfetchedPhrasesPrefaces(fetchTimestamp time.Time) (map[string]int64, map[string]int64, error) {
	phrases := map[string]int64{}
	rows, err := db.db.Query("SELECT phraseID, phrase FROM phrases WHERE lastFetchTimestamp < ?", fetchTimestamp.Format(time.DateTime))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var phraseID int64
		var phrase string
		if err := rows.Scan(&phraseID, &phrase); err != nil {
			return nil, nil, err
		}
		phrases[phrase] = phraseID
	}

	prefaces := map[string]int64{}
	rows, err = db.db.Query("SELECT prefaceID, preface FROM prefaces WHERE lastFetchTimestamp < ?", fetchTimestamp.Format(time.DateTime))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var prefaceID int64
		var preface string
		if err := rows.Scan(&prefaceID, &preface); err != nil {
			return nil, nil, err
		}
		prefaces[preface] = prefaceID
	}

	return phrases, prefaces, nil
}

func (db *phraseDB) addPhrase(phrase string) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT INTO phrases (phrase, lastFetchTimestamp) VALUES (?, '1970-01-01 00:00:00')", phrase); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *phraseDB) addPreface(preface string) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT INTO prefaces (preface, lastFetchTimestamp) VALUES (?, '1970-01-01 00:00:00')", preface); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *phraseDB) addCounts(fileID int64, date time.Time, phrases, prefaces map[string]int64, phraseCounts map[[2]string]int, prefaceCounts map[[2]string]int) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT OR IGNORE INTO files (fileID, date) VALUES (?,?)", fileID, date.Format(time.DateOnly)); err != nil {
		return err
	}

	speakerIDs := map[string]int64{}
	for item := range phraseCounts {
		if _, ok := speakerIDs[item[0]]; ok {
			continue
		}
		speakerID, err := db.getSpeakerID(tx, item[0])
		if err != nil {
			return err
		}
		speakerIDs[item[0]] = speakerID
	}
	for item := range prefaceCounts {
		if _, ok := speakerIDs[item[0]]; ok {
			continue
		}
		speakerID, err := db.getSpeakerID(tx, item[0])
		if err != nil {
			return err
		}
		speakerIDs[item[0]] = speakerID
	}

	for item, count := range phraseCounts {
		phraseID, ok := phrases[item[1]]
		if !ok || count <= 0 {
			continue
		}
		if _, err := tx.Exec("INSERT INTO phraseCounts (fileID, speakerID, phraseID, count) VALUES (?,?,?,?)", fileID, speakerIDs[item[0]], phraseID, count); err != nil {
			return nil
		}
	}

	for item, count := range prefaceCounts {
		prefaceID, ok := prefaces[item[1]]
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
