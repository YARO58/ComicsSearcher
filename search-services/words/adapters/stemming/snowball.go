package stemming

import (
	"strings"

	"github.com/kljensen/snowball/english"
)

type Snowball struct{}

func (s Snowball) Stem(word string) string {
	word = strings.ToLower(word)
	if english.IsStopWord(word) {
		return ""
	}
	return english.Stem(word, false)
}
