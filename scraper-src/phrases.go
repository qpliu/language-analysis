package scraper

import (
	"strings"
	"unicode"
)

type Phraser struct {
	words []string

	currentPhrase     []string
	currentIndex      int
	currentTerminated bool
}

func MakePhraser(maxWords int, text string) *Phraser {
	return &Phraser{
		words:         strings.Split(text, " "),
		currentPhrase: make([]string, maxWords),
	}
}

func (p *Phraser) OnlyFirstWords(count int) {
	if len(p.words) > count {
		p.words = p.words[:count]
	}
}

func (p *Phraser) phrase() []string {
	phrase := make([]string, len(p.currentPhrase))
	copy(phrase, p.currentPhrase)
	return phrase
}

func (p *Phraser) pushWord(word string) {
	copy(p.currentPhrase, p.currentPhrase[1:])
	p.currentPhrase[len(p.currentPhrase)-1] = word
}

func (p *Phraser) Next() []string {
	if p.currentIndex == 0 {
		if len(p.words) == 0 {
			return nil
		}
		p.currentTerminated = false
	}
	if p.currentTerminated {
		p.pushWord("")
		p.currentIndex--
		if p.currentIndex > 0 {
			return p.phrase()
		}
		p.currentTerminated = false
	}
	changedCurrent := false
	for len(p.words) > 0 {
		word := p.words[0]
		p.words = p.words[1:]
		if word == "" {
			continue
		}
		switch word[len(word)-1] {
		case '.', ':', ')', ']', '}', '!', '?', '"', '\'':
			p.currentTerminated = true
		}
		word = strings.ToLower(strings.TrimFunc(word, trimFunc))
		if word == "" {
			if p.currentTerminated {
				if p.currentIndex > 0 {
					if changedCurrent {
						return p.phrase()
					}
					p.pushWord("")
					changedCurrent = true
					p.currentIndex--
					if p.currentIndex > 0 {
						return p.phrase()
					}
				}
				p.currentTerminated = false
			}
			continue
		}
		changedCurrent = true
		if p.currentIndex == len(p.currentPhrase) {
			p.pushWord(word)
			return p.phrase()
		}
		p.currentPhrase[p.currentIndex] = word
		p.currentIndex++
		if p.currentIndex == len(p.currentPhrase) || p.currentTerminated {
			return p.phrase()
		}
	}
	p.currentTerminated = true
	if changedCurrent {
		return p.phrase()
	}
	p.pushWord("")
	p.currentIndex--
	if p.currentIndex > 0 {
		return p.phrase()
	}
	return nil
}

func trimFunc(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsNumber(r) {
		return false
	}
	return true
}
