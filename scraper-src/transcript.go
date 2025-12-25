package scraper

import (
	"fmt"
	"strings"
)

type Transcript struct {
	Index   int
	Speaker string
	Name    string
	Text    string
}

func (t Transcript) String() string {
	if t.Speaker != "" {
		if t.Name != "" && t.Name != t.Speaker {
			return fmt.Sprintf("[%s] %s: %s", t.Name, t.Speaker, t.Text)
		} else {
			return fmt.Sprintf("%s: %s", t.Speaker, t.Text)
		}
	} else {
		return t.Text
	}
}

func toTranscript(text []string) []Transcript {
	names := map[string]string{}
	transcript := []Transcript{}
	for index, item := range text {
		i := strings.Index(item, ":")
		if i < 0 {
			transcript = append(transcript, Transcript{
				Index: index,
				Text:  item,
			})
			continue
		}
		speaker := item[:i]
		text := strings.Trim(item[i+1:], " ")
		name := speaker
		comma := strings.Index(name, ",")
		if comma > 0 {
			name = strings.Trim(name[:comma], " ")
			names[name] = name
		} else if n, ok := names[name]; ok {
			name = n
		} else {
			found := false
			for _, n := range names {
				if strings.HasSuffix(n, name) {
					names[name] = n
					name = n
					found = true
					break
				}
			}
			if !found {
				for _, n := range names {
					if strings.HasPrefix(n, name) {
						names[name] = n
						name = n
						found = true
						break
					}
				}
			}
			if !found {
				split := strings.SplitN(name, " ", 2)
				if len(split) == 2 {
					for _, n := range names {
						if strings.HasPrefix(n, split[0]) && strings.HasSuffix(n, split[1]) {
							names[name] = n
							name = n
							found = true
							break
						}
					}
				}
			}
			if !found {
				names[name] = name
			}
		}
		transcript = append(transcript, Transcript{
			Index:   index,
			Speaker: speaker,
			Name:    name,
			Text:    text,
		})
	}
	return transcript
}
