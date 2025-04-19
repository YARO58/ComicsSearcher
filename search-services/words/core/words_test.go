package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockStemmer struct {
	stemFunc func(word string) string
}

func (m MockStemmer) Stem(word string) string {
	if m.stemFunc != nil {
		return m.stemFunc(word)
	}
	return word
}

func TestWords_Norm(t *testing.T) {
	tests := []struct {
		name     string
		phrase   string
		stemFunc func(string) string
		want     []string
	}{
		{
			name:   "empty string",
			phrase: "",
			want:   []string{},
		},
		{
			name:   "simple words",
			phrase: "hello world",
			want:   []string{"hello", "world"},
		},
		{
			name:   "with punctuation",
			phrase: "hello, world!",
			want:   []string{"hello", "world"},
		},
		{
			name:   "with numbers",
			phrase: "test123 456test",
			want:   []string{"test123", "456test"},
		},
		{
			name:   "with stemming",
			phrase: "running jumps",
			stemFunc: func(word string) string {
				if word == "running" {
					return "run"
				}
				if word == "jumps" {
					return "jump"
				}
				return word
			},
			want: []string{"run", "jump"},
		},
		{
			name:   "duplicate words",
			phrase: "hello hello world world",
			want:   []string{"hello", "world"},
		},
		{
			name:   "mixed case",
			phrase: "Hello World",
			want:   []string{"Hello", "World"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stemmer := MockStemmer{stemFunc: tt.stemFunc}
			words := NewWords(stemmer)
			got := words.Norm(tt.phrase)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
