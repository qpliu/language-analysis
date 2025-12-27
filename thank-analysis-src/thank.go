package thankAnalysis

import (
	"regexp"

	scraper "language-analysis/scraper-src"
)

const MaxWords = 5

var thanksRegex = regexp.MustCompile(`\b[Tt]hank(s| you)\b`)

func ThankResponses(transcript []scraper.Transcript) []scraper.Transcript {
	resps := map[string]scraper.Transcript{}
	lastWasThanks := false
	saidThanks := map[string]bool{}
	for _, ts := range transcript {
		if lastWasThanks && ts.Name != "" && !saidThanks[ts.Name] {
			resps[ts.Name] = ts
		} else {
			delete(resps, ts.Name)
		}
		lastWasThanks = thanksRegex.MatchString(ts.Text)
		saidThanks[ts.Name] = lastWasThanks
	}
	results := []scraper.Transcript{}
	for _, ts := range resps {
		results = append(results, ts)
	}
	return results
}

func ResponsePhrases(text string) map[[MaxWords]string]bool {
	phrases := map[[MaxWords]string]bool{}

	phrase := [MaxWords]string{}
	phraser := scraper.MakePhraser(MaxWords, text)
	phraser.OnlyFirstWords(20)
	for p := phraser.Next(); p != nil; p = phraser.Next() {
		for i := range MaxWords {
			phrase[i] = ""
		}
		copy(phrase[:], p)
		phrases[phrase] = true
		for i := range MaxWords - 1 {
			phrase[MaxWords-1-i] = ""
			phrases[phrase] = true
		}
	}
	return phrases
}
