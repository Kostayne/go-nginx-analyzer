package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDateTime(t *testing.T) {
	t.Run("should parse date with square brackets", func(t *testing.T) {
		dateStr := "[25/Dec/2023:10:30:45 +0000]"

		result, err := parseDateTime(dateStr)

		assert.NoError(t, err)
		expected, _ := time.Parse("02/Jan/2006:15:04:05 -0700", "25/Dec/2023:10:30:45 +0000")
		assert.Equal(t, expected, result)
	})

	t.Run("should parse date without square brackets", func(t *testing.T) {
		dateStr := "25/Dec/2023:10:30:45 +0000"

		result, err := parseDateTime(dateStr)

		assert.NoError(t, err)
		expected, _ := time.Parse("02/Jan/2006:15:04:05 -0700", "25/Dec/2023:10:30:45 +0000")
		assert.Equal(t, expected, result)
	})

	t.Run("should return error for invalid date format", func(t *testing.T) {
		dateStr := "invalid-date"

		result, err := parseDateTime(dateStr)

		assert.Error(t, err)
		assert.Equal(t, time.Time{}, result)
	})

	t.Run("should return error for empty date string", func(t *testing.T) {
		dateStr := ""

		result, err := parseDateTime(dateStr)

		assert.Error(t, err)
		assert.Equal(t, time.Time{}, result)
	})
}
