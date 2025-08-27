package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseLogEntry(t *testing.T) {
	t.Run("should parse valid nginx access log entry correctly", func(t *testing.T) {
		line := `192.168.1.100 - - [25/Dec/2023:10:30:45 +0000] "GET /api/users HTTP/1.1" 200 1234 "https://example.com" "Mozilla/5.0"`

		log, err := ParseLogEntry(line)

		assert.NoError(t, err)
		assert.NotNil(t, log)
		assert.Equal(t, "192.168.1.100", log.Ip)
		assert.Equal(t, "-", log.User)
		assert.Equal(t, `"GET`, log.Method) // Quotes remain in method field
		assert.Equal(t, "/api/users", log.Uri)
		assert.Equal(t, `HTTP/1.1"`, log.Protocol) // Quotes remain in protocol field
		assert.Equal(t, 200, log.StatusCode)
		assert.Equal(t, 1234, log.RespBytes)
		assert.Equal(t, `"https://example.com"`, log.Referrer) // Quotes remain in referrer
		assert.Equal(t, `"Mozilla/5.0"`, log.UserAgent)        // Quotes remain in user agent

		// Check date parsing
		expectedTime, _ := time.Parse("02/Jan/2006:15:04:05 -0700", "25/Dec/2023:10:30:45 +0000")
		assert.Equal(t, expectedTime, log.Date)
	})

	t.Run("should return error when status code is not a number", func(t *testing.T) {
		line := `192.168.1.100 - - [25/Dec/2023:10:30:45 +0000] "GET /api/users HTTP/1.1" abc 1234 "https://example.com" "Mozilla/5.0"`

		log, err := ParseLogEntry(line)

		assert.Error(t, err)
		assert.Nil(t, log)
	})

	t.Run("should return error when response bytes is not a number", func(t *testing.T) {
		line := `192.168.1.100 - - [25/Dec/2023:10:30:45 +0000] "GET /api/users HTTP/1.1" 200 abc "https://example.com" "Mozilla/5.0"`

		log, err := ParseLogEntry(line)

		assert.Error(t, err)
		assert.Nil(t, log)
	})

	t.Run("should return error when date format is invalid", func(t *testing.T) {
		line := `192.168.1.100 - - [invalid-date] "GET /api/users HTTP/1.1" 200 1234 "https://example.com" "Mozilla/5.0"`

		log, err := ParseLogEntry(line)

		assert.Error(t, err)
		assert.Nil(t, log)
	})

	t.Run("should return error when line is empty", func(t *testing.T) {
		line := ""

		log, err := ParseLogEntry(line)

		assert.Error(t, err) // Empty line should cause date parsing error
		assert.Nil(t, log)
	})
}
