package stemming

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnowball_Stem(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "stop word",
			input:    "the",
			expected: "",
		},
		{
			name:     "simple word",
			input:    "running",
			expected: "run",
		},
		{
			name:     "uppercase word",
			input:    "RUNNING",
			expected: "run",
		},
		{
			name:     "mixed case word",
			input:    "RuNnInG",
			expected: "run",
		},
		{
			name:     "word with numbers",
			input:    "test123",
			expected: "test123",
		},
		{
			name:     "word with special characters",
			input:    "test!@#",
			expected: "test!@#",
		},
		{
			name:     "plural word",
			input:    "cats",
			expected: "cat",
		},
		{
			name:     "word with -ing",
			input:    "jumping",
			expected: "jump",
		},
		{
			name:     "word with -ed",
			input:    "jumped",
			expected: "jump",
		},
	}

	stemmer := Snowball{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stemmer.Stem(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
