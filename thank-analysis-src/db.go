package thankAnalysis

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type thankDB struct {
	db *sql.DB
}

func openThankDB(dir string) (*thankDB, error) {
	db, err := sql.Open("sqlite3", dir+"/thank-analysis.db")
	if err != nil {
		return nil, err
	}

	tdb := thankDB{db}
	if err := tdb.init(); err != nil {
		tdb.Close()
		return nil, err
	}
	return &tdb, nil
}

func (db *thankDB) Close() error {
	return db.db.Close()
}

func (db *thankDB) init() error {
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
		`CREATE TABLE words (
			wordID INTEGER PRIMARY KEY AUTOINCREMENT,
			word TEXT UNIQUE)`,
		`CREATE TABLE speakers (
			speakerID INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE)`,
		`CREATE INDEX speakersName ON speakers (name)`,
		`CREATE TABLE files (
			fileID INTEGER PRIMARY KEY,
			date DATE)`,
		`CREATE INDEX fileDate ON files (date)`,
		`CREATE TABLE responses (
			fileID INTEGER REFERENCES files (fileID),
			speakerID INTEGER REFERENCES speakers (speakerID),
			word1ID INTEGER REFERENCES words (wordID),
			word2ID INTEGER REFERENCES words (wordID),
			word3ID INTEGER REFERENCES words (wordID),
			word4ID INTEGER REFERENCES words (wordID),
			word5ID INTEGER REFERENCES words (wordID),
			PRIMARY KEY (fileID, speakerID,
				word1ID, word2ID, word3ID, word4ID, word5ID))`,
		`CREATE INDEX responsesWords
			ON responses (word1ID, word2ID, word3ID,
					word4ID, word5ID)`,
	} {
		if _, err := tx.Exec(statement); err != nil {
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

func (db *thankDB) lastFetchTimestamp() (time.Time, error) {
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

func (db *thankDB) setFetchTimestamp(lastFetchTimestamp time.Time) error {
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

func (db *thankDB) addResponses(fileID int64, date time.Time, responses map[string]map[[5]string]bool) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT INTO files (fileID, date) VALUES (?,?)", fileID, date.Format(time.DateOnly)); err != nil {
		return err
	}

	wordIDs := map[string]int64{}
	speakerWordIDs := [][6]int64{}
	for speaker, responsePhrases := range responses {
		speakerID, err := db.getSpeakerID(tx, speaker)
		if err != nil {
			return err
		}
		for responsePhrase := range responsePhrases {
			speakerWordID, err := db.getSpeakerWordID(tx, speakerID, responsePhrase, wordIDs)
			if err != nil {
				return err
			}
			speakerWordIDs = append(speakerWordIDs, speakerWordID)
		}
	}

	for _, speakerWordID := range speakerWordIDs {
		if _, err := tx.Exec("INSERT INTO responses (fileID, speakerID, word1ID, word2ID, word3ID, word4ID, word5ID) VALUES (?,?,?,?,?,?,?)", fileID, speakerWordID[0], speakerWordID[1], speakerWordID[2], speakerWordID[3], speakerWordID[4], speakerWordID[5]); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *thankDB) getSpeakerID(tx *sql.Tx, speaker string) (int64, error) {
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

func (db *thankDB) getSpeakerWordID(tx *sql.Tx, speakerID int64, responsePhrase [5]string, wordIDs map[string]int64) ([6]int64, error) {
	speakerWordID := [6]int64{speakerID, 0, 0, 0, 0, 0}
	for i, word := range responsePhrase {
		wordID, err := db.getWordID(tx, word, wordIDs)
		if err != nil {
			return [6]int64{}, err
		}
		speakerWordID[i+1] = wordID
	}
	return speakerWordID, nil
}

func (db *thankDB) getWordID(tx *sql.Tx, word string, wordIDs map[string]int64) (int64, error) {
	if wordID, ok := wordIDs[word]; ok {
		return wordID, nil
	}
	for range 2 {
		rows, err := tx.Query("SELECT wordID FROM words WHERE word = ?", word)
		if err != nil {
			return 0, err
		}

		for rows.Next() {
			var wordID int64
			if err := rows.Scan(&wordID); err != nil {
				return 0, err
			}
			wordIDs[word] = wordID
			return wordID, nil
		}

		result, err := tx.Exec("INSERT INTO words (word) VALUES (?)", word)
		if err != nil {
			return 0, err
		}

		if wordID, err := result.LastInsertId(); err == nil {
			wordIDs[word] = wordID
			return wordID, nil
		}
	}
	return 0, fmt.Errorf("Failed to get wordID for %s", word)
}
