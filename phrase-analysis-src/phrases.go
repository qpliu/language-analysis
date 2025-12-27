package phraseAnalysis

import (
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
		phraser := scraper.MakePhraser(maxWords, ts.Text)
		preface := true
		for phrase := phraser.Next(); phrase != nil; phrase = phraser.Next() {
			collectPhrase(ts.Name, phrase, phraseCounts)
			if preface {
				collectPreface(ts.Name, phrase, prefaceCounts)
				preface = false
			}
		}
	}
	return phraseCounts, prefaceCounts
}

func collectPhrase(speaker string, phrase []string, phraseCounts map[[2]string]int) {
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

func collectPreface(speaker string, phrase []string, prefaceCounts map[[2]string]int) {
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
