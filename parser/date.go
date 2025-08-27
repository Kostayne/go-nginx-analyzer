package parser

import (
	"strings"
	"time"
)

func parseDateTime(dateStr string) (time.Time, error) {
	str := strings.Trim(dateStr, "[]")

	parsedTime, err := time.Parse("02/Jan/2006:15:04:05 -0700", str)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}
