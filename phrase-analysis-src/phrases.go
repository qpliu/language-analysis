package phraseAnalysis

import (
	"strings"
	"unicode"

	scraper "language-analysis/scraper-src"
)

const maxWords = 5

var (
	PHRASES = [...]string{
		"bucket list",
		"perfect storm",
		"absolutely",
		"definitely",
		"exponentially",
		"you know",
	}

	PREFACES = [...]string{
		"absolutely",
		"look",
	}

	PHRASEIDS  map[string]int
	PREFACEIDS map[string]int
)

func init() {
	PHRASEIDS = map[string]int{}
	for i, phrase := range PHRASES {
		PHRASEIDS[phrase] = i
	}
	PREFACEIDS = map[string]int{}
	for i, preface := range PREFACES {
		PREFACEIDS[preface] = i
	}
}

func CountPhrases(transcript []scraper.Transcript) (map[[2]string]int, map[[2]string]int) {
	phraseCounts := map[[2]string]int{}
	prefaceCounts := map[[2]string]int{}

	for _, ts := range transcript {
		phrase := [maxWords]string{}
		preface := true
		for _, word := range strings.Split(ts.Text, " ") {
			word = strings.ToLower(strings.TrimFunc(word, trimFunc))
			for i := range maxWords {
				if phrase[i] == "" {
					phrase[i] = word
					break
				} else if i == maxWords-1 {
					collectPhrase(ts.Name, phrase, phraseCounts)
					if preface {
						collectPreface(ts.Name, phrase, prefaceCounts)
						preface = false
					}
					for j := range i {
						phrase[j] = phrase[j+1]
					}
					phrase[i] = word
				}
			}
		}
		for phrase[0] != "" {
			collectPhrase(ts.Name, phrase, phraseCounts)
			if preface {
				collectPreface(ts.Name, phrase, prefaceCounts)
				preface = false
			}
			for j := range maxWords - 1 {
				phrase[j] = phrase[j+1]
			}
			phrase[maxWords-1] = ""
		}
	}
	return phraseCounts, prefaceCounts
}

func trimFunc(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsNumber(r) {
		return false
	}
	return true
}

func collectPhrase(speaker string, phrase [maxWords]string, phraseCounts map[[2]string]int) {
	p := phrase[0]
	if _, ok := PHRASEIDS[p]; ok {
		phraseCounts[[2]string{speaker, p}]++
	}
	for _, w := range phrase[1:] {
		if w != "" {
			p += " " + w
			if _, ok := PHRASEIDS[p]; ok {
				phraseCounts[[2]string{speaker, p}]++
			}
		}
	}
}

func collectPreface(speaker string, phrase [maxWords]string, prefaceCounts map[[2]string]int) {
	p := phrase[0]
	if _, ok := PREFACEIDS[p]; ok {
		prefaceCounts[[2]string{speaker, p}]++
	}
	for _, w := range phrase[1:] {
		if w != "" {
			p += " " + w
			if _, ok := PREFACEIDS[p]; ok {
				prefaceCounts[[2]string{speaker, p}]++
			}
		}
	}
}
