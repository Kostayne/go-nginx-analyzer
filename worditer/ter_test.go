package worditer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordIter(t *testing.T) {
	t.Run("should return empty string and false when no more words available", func(t *testing.T) {
		iter := New("")
		word, hasMore := iter.Next()

		assert.Equal(t, word, "")
		assert.False(t, hasMore)
	})

	t.Run("should return empty string when no more words available using NextOrEmpty", func(t *testing.T) {
		iter := New("")
		word := iter.NextOrEmpty()

		assert.Equal(t, word, "")
	})

	t.Run("should return words sequentially with Next method", func(t *testing.T) {
		iter := New("Hello world!")

		word, hasMore := iter.Next()
		assert.Equal(t, word, "Hello")
		assert.True(t, hasMore)

		word, hasMore = iter.Next()
		assert.Equal(t, word, "world!")
		assert.False(t, hasMore)
	})

	t.Run("should return words sequentially with NextOrEmpty method", func(t *testing.T) {
		iter := New("Hello world!")

		word := iter.NextOrEmpty()
		assert.Equal(t, word, "Hello")

		word = iter.NextOrEmpty()
		assert.Equal(t, word, "world!")
	})
}
