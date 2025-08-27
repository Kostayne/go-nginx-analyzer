package worditer

import (
	"slices"
	"strings"
)

type WordIter struct {
	words []string
}

func (iter *WordIter) HasMore() bool {
	return len(iter.words) > 0 && iter.words[0] != ""
}

// Returns a word and boolean value indicating if the word was the last one
func (iter *WordIter) Next() (string, bool) {
	if !iter.HasMore() {
		return "", false
	}

	word := iter.words[0]
	iter.words = iter.words[1:]

	return word, iter.HasMore()
}

func (iter *WordIter) NextOrEmpty() string {
	if !iter.HasMore() {
		return ""
	}

	word := iter.words[0]
	iter.words = iter.words[1:]

	return word
}

func New(line string) *WordIter {
	words := strings.Split(line, " ")
	words = slices.DeleteFunc(words, func(word string) bool {
		return word == ""
	})

	return &WordIter{words: words}
}
