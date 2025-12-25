package thankAnalysis

import (
	"regexp"
	"strings"
	"unicode"

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

func ResponsePhrases(text string) map[[5]string]bool {
	phrases := map[[5]string]bool{}

	phrase := [5]string{}
	words := strings.Split(text, " ")
	for range 20 {
		if len(words) == 0 {
			break
		}
		word := strings.ToLower(strings.TrimFunc(words[0], trimFunc))
		words = words[1:]
		if word == "" {
			continue
		}
		if phrase[0] == "" {
			phrase[0] = word
		} else if phrase[1] == "" {
			phrase[1] = word
		} else if phrase[2] == "" {
			phrase[2] = word
		} else if phrase[3] == "" {
			phrase[3] = word
		} else if phrase[4] == "" {
			phrase[4] = word
		} else {
			phrase[0], phrase[1], phrase[2], phrase[3], phrase[4] = phrase[1], phrase[2], phrase[3], phrase[4], word
		}
		if phrase[0] != "" {
			p := phrase
			phrases[p] = true
			p[4] = ""
			phrases[p] = true
			p[3] = ""
			phrases[p] = true
			p[2] = ""
			phrases[p] = true
			p[1] = ""
			phrases[p] = true
		}
	}
	for phrase[0] != "" {
		p := phrase
		phrases[p] = true
		p[4] = ""
		phrases[p] = true
		p[3] = ""
		phrases[p] = true
		p[2] = ""
		phrases[p] = true
		p[1] = ""
		phrases[p] = true
		phrase[0], phrase[1], phrase[2], phrase[3], phrase[4] = phrase[1], phrase[2], phrase[3], phrase[4], ""
	}
	return phrases
}

func trimFunc(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsNumber(r) {
		return false
	}
	return true
}
