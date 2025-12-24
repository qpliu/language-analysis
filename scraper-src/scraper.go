package scraper

import (
	"bytes"
	"regexp"

	fetcher "language-analysis/fetcher-src"
)

var startMarker = regexp.MustCompile(`<div class="[^"]*transcript[^"]*"[^>]*>`)
var contentMarker = regexp.MustCompile(`<p>`)
var endMarker = regexp.MustCompile(`</div>`)

func Scrape(file fetcher.File) ([]Transcript, error) {
	data, err := file.Contents()
	if err != nil {
		return nil, err
	}
	return toTranscript(scrapeContents(data)), nil
}

func scrapeContents(data []byte) []string {
	loc := startMarker.FindIndex(data)
	if loc == nil {
		return nil
	}
	data = data[loc[1]:]
	loc = endMarker.FindIndex(data)
	if loc != nil {
		data = data[:loc[0]]
	}

	loc = contentMarker.FindIndex(data)
	if loc == nil {
		return nil
	}
	data = data[loc[1]:]

	items := []string{}
	for {
		loc := contentMarker.FindIndex(data)
		if loc == nil {
			return items
		}
		if loc[0] > 0 {
			buf := bytes.Buffer{}
			buf.Write(data[:loc[0]])
			items = append(items, buf.String())
		}
		data = data[loc[1]:]
	}
}
