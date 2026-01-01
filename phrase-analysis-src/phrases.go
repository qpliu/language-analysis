package phraseAnalysis

import (
	"time"

	scraper "language-analysis/scraper-src"
)

const maxWords = 5

func PhrasesPrefaces() (map[string]int64, map[string]int64, error) {
	db, err := openPhraseDB()
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	return db.unfetchedPhrasesPrefaces(time.Now().UTC())
}

func AddPhrase(phrase string) error {
	db, err := openPhraseDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.addPhrase(phrase)
}

func AddPreface(preface string) error {
	db, err := openPhraseDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.addPreface(preface)
}

func CountPhrases(transcript []scraper.Transcript, phrases, prefaces map[string]int64) (map[[2]string]int, map[[2]string]int) {
	phraseCounts := map[[2]string]int{}
	prefaceCounts := map[[2]string]int{}

	for _, ts := range transcript {
		phraser := scraper.MakePhraser(maxWords, ts.Text)
		preface := true
		for phrase := phraser.Next(); phrase != nil; phrase = phraser.Next() {
			collect(phrases, ts.Name, phrase, phraseCounts)
			if preface {
				collect(prefaces, ts.Name, phrase, prefaceCounts)
				preface = false
			}
		}
	}
	return phraseCounts, prefaceCounts
}

func collect(phrases map[string]int64, speaker string, phrase []string, phraseCounts map[[2]string]int) {
	p := phrase[0]
	if _, ok := phrases[p]; ok {
		phraseCounts[[2]string{speaker, p}]++
	}
	for _, w := range phrase[1:] {
		if w != "" {
			p += " " + w
			if _, ok := phrases[p]; ok {
				phraseCounts[[2]string{speaker, p}]++
			}
		}
	}
}
