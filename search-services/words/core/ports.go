package core

// our logic
type Normalizer interface {
	Norm(phrase string) []string
}

// external stemmer
type Stemmer interface {
	Stem(word string) string
}
