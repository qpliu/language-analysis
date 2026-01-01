package fetcher

import (
	"compress/gzip"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"language-analysis/config"
)

type File struct {
	fileID int64
	feedID int64
	url    string
	date   time.Time

	fetchTimestamp time.Time
	purgeTimestamp time.Time
}

func (file File) Filename() string {
	dirname, filename := filename(file)
	return dirname + "/" + filename
}

func (file File) Date() time.Time {
	return file.date
}

func (file File) FetchTimestamp() time.Time {
	return file.fetchTimestamp
}

func (file File) PurgeTimestamp() time.Time {
	return file.purgeTimestamp
}

func (file File) ID() int64 {
	return file.fileID
}

func (file File) Contents() ([]byte, error) {
	fd, err := os.Open(file.Filename())
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	in, err := gzip.NewReader(fd)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	return io.ReadAll(in)
}

func FilesSince(since time.Time, limit int) ([]File, error) {
	db, err := openFetcherDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return db.fetchedSince(since, limit)
}

func FileByID(fileID int64) (File, error) {
	db, err := openFetcherDB()
	if err != nil {
		return File{}, err
	}
	defer db.Close()

	return db.file(fileID)
}

func fetchFile(file File, db *fetcherDB) error {
	fileData, err := fetch(file.url)
	if err != nil {
		return err
	}

	dirname, filename := filename(file)
	if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
		return err
	}

	fd, err := os.Create(dirname + "/" + filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	gz := gzip.NewWriter(fd)
	defer gz.Close()
	if _, err := gz.Write(fileData); err != nil {
		return err
	}

	if err := db.updateFileFetched(file.fileID); err != nil {
		return err
	}

	return nil
}

func filename(file File) (string, string) {
	hash := sha1.Sum([]byte(file.url))
	return fmt.Sprintf("%s/files/%x/%x", config.Dir(), hash[0:2], hash[2:4]), base64.RawURLEncoding.EncodeToString([]byte(file.url))
}
