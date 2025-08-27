package parser

import (
	"fmt"
	"net/netip"
	"strconv"
	"time"

	"github.com/Kostayne/go-nginx-analyzer/worditer"
)

type LogEntry struct {
	Ip         netip.Addr
	User       string
	Date       time.Time
	Method     string
	Uri        string
	Protocol   string
	StatusCode uint16
	RespBytes  uint
	Referrer   string
	UserAgent  string
}

func ParseLogEntry(line string) (*LogEntry, error) {
	log := &LogEntry{}
	iter := worditer.New(line)

	ipStr := iter.NextOrEmpty()
	if ipStr == "" {
		return nil, fmt.Errorf("ip is empty")
	}

	parsedIp, err := netip.ParseAddr(ipStr)
	log.Ip = parsedIp
	if err != nil {
		return nil, err
	}

	iter.NextOrEmpty()

	log.User = iter.NextOrEmpty()
	if log.User == "" {
		return nil, fmt.Errorf("user is empty")
	}

	dateStr := fmt.Sprintf("%s %s", iter.NextOrEmpty(), iter.NextOrEmpty())
	parsedDate, err := parseDateTime(dateStr)
	if err != nil {
		return nil, err
	}
	log.Date = parsedDate

	log.Method = iter.NextOrEmpty()
	if log.Method == "" {
		return nil, fmt.Errorf("method is empty")
	}

	log.Uri = iter.NextOrEmpty()
	if log.Uri == "" {
		return nil, fmt.Errorf("uri is empty")
	}

	log.Protocol = iter.NextOrEmpty()
	if log.Protocol == "" {
		return nil, fmt.Errorf("protocol is empty")
	}

	statusCode, err := strconv.ParseUint(iter.NextOrEmpty(), 10, 16)
	log.StatusCode = uint16(statusCode)
	if err != nil {
		return nil, err
	}

	respBytes, err := strconv.ParseUint(iter.NextOrEmpty(), 10, 32)
	log.RespBytes = uint(respBytes)
	if err != nil {
		return nil, err
	}

	log.Referrer = iter.NextOrEmpty()
	if log.Referrer == "" {
		return nil, fmt.Errorf("referrer is empty")
	}

	log.UserAgent = iter.NextOrEmpty()
	if log.UserAgent == "" {
		return nil, fmt.Errorf("user agent is empty")
	}

	return log, nil
}
