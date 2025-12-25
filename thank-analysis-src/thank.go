package thankAnalysis

import (
	"regexp"

	scraper "language-analysis/scraper-src"
)

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
