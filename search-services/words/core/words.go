package core

import (
	"maps"
	"slices"
	"strings"
	"unicode"

	"log/slog"
)

type Words struct {
	stemmer Stemmer
}

func NewWords(stemmer Stemmer) Words {
	return Words{stemmer: stemmer}
}

func (w Words) Norm(phrase string) []string {

	splitted := strings.FieldsFunc(phrase, func(r rune) bool {
		return !unicode.IsDigit(r) && !unicode.IsLetter(r)
	})

	words := make(map[string]bool)
	for _, word := range splitted {
		stemmed := w.stemmer.Stem(word)
		if len(stemmed) > 0 {
			words[stemmed] = true
		}
	}

	slog.Info("words normalized", "words", words)

	return slices.Collect(maps.Keys(words))
}
