package analyzer

import (
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyze(t *testing.T) {
	t.Run("should analyze valid log file successfully", func(t *testing.T) {
		// Create a temporary test file with valid log entries
		testData := `192.168.1.100 - - [25/Dec/2023:10:30:45 +0000] "GET /api/users HTTP/1.1" 200 1234 "https://example.com" "Mozilla/5.0"
192.168.1.101 - - [25/Dec/2023:10:31:45 +0000] "POST /api/users HTTP/1.1" 201 567 "https://example.com" "Mozilla/5.0"
192.168.1.100 - - [25/Dec/2023:10:32:45 +0000] "GET /api/posts HTTP/1.1" 404 123 "https://example.com" "Mozilla/5.0"
192.168.1.102 - - [25/Dec/2023:10:33:45 +0000] "GET /api/users HTTP/1.1" 200 1234 "https://example.com" "Chrome/5.0"`

		tmpFile, err := os.CreateTemp("", "test_log_*.log")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.WriteString(testData)
		require.NoError(t, err)

		result, err := Analyze(tmpFile.Name(), 10, true, "hour")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint64(4), result.TotalRequests)
		assert.Equal(t, uint64(3), result.UniqueIPs)
		assert.Equal(t, uint64(2), result.UniqueUserAgents)
		assert.Equal(t, uint64(0), result.ProcessingStats.ParseErrors)

		// Check IP hits
		assert.Len(t, result.Ips, 3)
		assert.Equal(t, "192.168.1.100", result.Ips[0].Key.String())
		assert.Equal(t, uint64(2), result.Ips[0].Hits)

		// Check status codes
		assert.Len(t, result.Codes, 3)
		assert.Equal(t, uint16(200), result.Codes[0].Key)
		assert.Equal(t, uint64(2), result.Codes[0].Hits)

		// Check dates
		assert.Len(t, result.Dates, 1) // All entries in same hour
	})

	t.Run("should handle file with no valid log entries", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "no_valid_entries_log_*.log")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		// Write some invalid log entries
		_, err = tmpFile.WriteString("invalid entry 1\ninvalid entry 2\n")
		require.NoError(t, err)

		result, err := Analyze(tmpFile.Name(), 10, true, "hour")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint64(0), result.TotalRequests)
		assert.Equal(t, uint64(0), result.UniqueIPs)
		assert.Equal(t, uint64(0), result.UniqueUserAgents)
		assert.Equal(t, uint64(2), result.ProcessingStats.ParseErrors)
	})

	t.Run("should handle file with parse errors", func(t *testing.T) {
		testData := `192.168.1.100 - - [25/Dec/2023:10:30:45 +0000] "GET /api/users HTTP/1.1" 200 1234 "https://example.com" "Mozilla/5.0"
invalid log entry
192.168.1.101 - - [25/Dec/2023:10:31:45 +0000] "POST /api/users HTTP/1.1" 201 567 "https://example.com" "Mozilla/5.0"`

		tmpFile, err := os.CreateTemp("", "error_log_*.log")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.WriteString(testData)
		require.NoError(t, err)

		result, err := Analyze(tmpFile.Name(), 10, true, "hour")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint64(2), result.TotalRequests)
		assert.Equal(t, uint64(1), result.ProcessingStats.ParseErrors)
	})

	t.Run("should return error for non-existent file", func(t *testing.T) {
		result, err := Analyze("non_existent_file.log", 10, true, "hour")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("should group by day correctly", func(t *testing.T) {
		testData := `192.168.1.100 - - [25/Dec/2023:10:30:45 +0000] "GET /api/users HTTP/1.1" 200 1234 "https://example.com" "Mozilla/5.0"
192.168.1.101 - - [25/Dec/2023:11:30:45 +0000] "POST /api/users HTTP/1.1" 201 567 "https://example.com" "Mozilla/5.0"
192.168.1.102 - - [26/Dec/2023:10:30:45 +0000] "GET /api/posts HTTP/1.1" 404 123 "https://example.com" "Mozilla/5.0"`

		tmpFile, err := os.CreateTemp("", "day_group_log_*.log")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.WriteString(testData)
		require.NoError(t, err)

		result, err := Analyze(tmpFile.Name(), 10, true, "day")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint64(3), result.TotalRequests)
		assert.Len(t, result.Dates, 2) // Two different days
	})

	t.Run("should sort in ascending order when desc is false", func(t *testing.T) {
		testData := `192.168.1.100 - - [25/Dec/2023:10:30:45 +0000] "GET /api/users HTTP/1.1" 200 1234 "https://example.com" "Mozilla/5.0"
192.168.1.101 - - [25/Dec/2023:10:31:45 +0000] "POST /api/users HTTP/1.1" 201 567 "https://example.com" "Mozilla/5.0"
192.168.1.100 - - [25/Dec/2023:10:32:45 +0000] "GET /api/posts HTTP/1.1" 404 123 "https://example.com" "Mozilla/5.0"`

		tmpFile, err := os.CreateTemp("", "asc_sort_log_*.log")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.WriteString(testData)
		require.NoError(t, err)

		result, err := Analyze(tmpFile.Name(), 10, false, "hour")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Ips, 2)
		// Should be sorted by hits in ascending order
		assert.Equal(t, uint64(1), result.Ips[0].Hits)
		assert.Equal(t, uint64(2), result.Ips[1].Hits)
	})
}

func TestGetHitsInfo(t *testing.T) {
	t.Run("should return top N hits in descending order", func(t *testing.T) {
		testMap := map[string]uint64{
			"a": 10,
			"b": 30,
			"c": 20,
			"d": 5,
		}

		result := getHitsInfo(testMap, 3, true)

		assert.Len(t, *result, 3)
		assert.Equal(t, "b", (*result)[0].Key)
		assert.Equal(t, uint64(30), (*result)[0].Hits)
		assert.Equal(t, "c", (*result)[1].Key)
		assert.Equal(t, uint64(20), (*result)[1].Hits)
		assert.Equal(t, "a", (*result)[2].Key)
		assert.Equal(t, uint64(10), (*result)[2].Hits)
	})

	t.Run("should return top N hits in ascending order", func(t *testing.T) {
		testMap := map[string]uint64{
			"a": 10,
			"b": 30,
			"c": 20,
			"d": 5,
		}

		result := getHitsInfo(testMap, 3, false)

		assert.Len(t, *result, 3)
		assert.Equal(t, "d", (*result)[0].Key)
		assert.Equal(t, uint64(5), (*result)[0].Hits)
		assert.Equal(t, "a", (*result)[1].Key)
		assert.Equal(t, uint64(10), (*result)[1].Hits)
		assert.Equal(t, "c", (*result)[2].Key)
		assert.Equal(t, uint64(20), (*result)[2].Hits)
	})

	t.Run("should handle empty map", func(t *testing.T) {
		testMap := map[string]uint64{}

		result := getHitsInfo(testMap, 10, true)

		assert.Len(t, *result, 0)
	})

	t.Run("should limit to available items when topN is larger", func(t *testing.T) {
		testMap := map[string]uint64{
			"a": 10,
			"b": 20,
		}

		result := getHitsInfo(testMap, 5, true)

		assert.Len(t, *result, 2)
	})
}

func TestProcessHitsInfo(t *testing.T) {
	t.Run("should process and sort hits correctly", func(t *testing.T) {
		hits := []HitsInfo[string]{
			{Key: "a", Hits: 10},
			{Key: "b", Hits: 30},
			{Key: "c", Hits: 20},
		}

		result := processHitsInfo(hits, 2, true)

		assert.Len(t, result, 2)
		assert.Equal(t, "b", result[0].Key)
		assert.Equal(t, uint64(30), result[0].Hits)
		assert.Equal(t, "c", result[1].Key)
		assert.Equal(t, uint64(20), result[1].Hits)
	})

	t.Run("should handle empty slice", func(t *testing.T) {
		hits := []HitsInfo[string]{}

		result := processHitsInfo(hits, 10, true)

		assert.Len(t, result, 0)
	})
}

func TestFindNewLineIndex(t *testing.T) {
	t.Run("should find newline index", func(t *testing.T) {
		data := []byte("hello\nworld")
		index := findNewLineIndex(data, 0)
		assert.Equal(t, 5, index)
	})

	t.Run("should return -1 when no newline found", func(t *testing.T) {
		data := []byte("hello world")
		index := findNewLineIndex(data, 0)
		assert.Equal(t, -1, index)
	})

	t.Run("should find newline after start position", func(t *testing.T) {
		data := []byte("hello\nworld\nend")
		index := findNewLineIndex(data, 6)
		assert.Equal(t, 11, index)
	})

	t.Run("should handle empty slice", func(t *testing.T) {
		data := []byte{}
		index := findNewLineIndex(data, 0)
		assert.Equal(t, -1, index)
	})
}

func TestNewChunk(t *testing.T) {
	t.Run("should create chunk with correct boundaries", func(t *testing.T) {
		fileName := "test.log"
		fileSize := int64(CHUNK_SIZE * 2) // Large enough file

		chunk := newChunk(0, fileName, fileSize)

		assert.Equal(t, fileName, chunk.fileName)
		assert.Equal(t, int64(0), chunk.startPos)
		// For index 0: start = 0*CHUNK_SIZE - CHUNK_OVERLAP = -CHUNK_OVERLAP, clamped to 0
		// end = 0 + CHUNK_SIZE + CHUNK_OVERLAP = CHUNK_SIZE + CHUNK_OVERLAP
		expectedEnd := int64(CHUNK_SIZE)
		assert.Equal(t, expectedEnd, chunk.endPos)
	})

	t.Run("should handle negative start position", func(t *testing.T) {
		fileName := "test.log"
		fileSize := int64(CHUNK_SIZE * 2) // Large enough file

		chunk := newChunk(1, fileName, fileSize)

		assert.Equal(t, fileName, chunk.fileName)
		// For index 1: start = 1*CHUNK_SIZE - CHUNK_OVERLAP = CHUNK_SIZE - CHUNK_OVERLAP
		expectedStart := int64(CHUNK_SIZE - CHUNK_OVERLAP)
		assert.Equal(t, expectedStart, chunk.startPos)
		// end = start + CHUNK_SIZE + CHUNK_OVERLAP = (CHUNK_SIZE - CHUNK_OVERLAP) + CHUNK_SIZE + CHUNK_OVERLAP = 2*CHUNK_SIZE
		expectedEnd := int64(2*CHUNK_SIZE - 1)
		assert.Equal(t, expectedEnd, chunk.endPos)
	})

	t.Run("should handle end position beyond file size", func(t *testing.T) {
		fileName := "test.log"
		fileSize := int64(1000) // Small file

		chunk := newChunk(10, fileName, fileSize)

		assert.Equal(t, fileName, chunk.fileName)
		// For index 10: start = 10*CHUNK_SIZE - CHUNK_OVERLAP = 10*CHUNK_SIZE - CHUNK_OVERLAP
		// end = start + CHUNK_SIZE + CHUNK_OVERLAP = 11*CHUNK_SIZE, but clamped to fileSize-1
		assert.Equal(t, int64(fileSize-1), chunk.endPos) // Should be clamped to fileSize-1
	})
}

func TestGetSortCompareResultAsc(t *testing.T) {
	t.Run("should return -1 when a < b", func(t *testing.T) {
		result := getSortCompareResultAsc(5, 10)
		assert.Equal(t, -1, result)
	})

	t.Run("should return 1 when a > b", func(t *testing.T) {
		result := getSortCompareResultAsc(10, 5)
		assert.Equal(t, 1, result)
	})

	t.Run("should return 0 when a == b", func(t *testing.T) {
		result := getSortCompareResultAsc(5, 5)
		assert.Equal(t, 0, result)
	})
}

func TestMergeResults(t *testing.T) {
	t.Run("should merge multiple results correctly", func(t *testing.T) {
		ip1, _ := netip.ParseAddr("192.168.1.100")
		ip2, _ := netip.ParseAddr("192.168.1.101")
		ip3, _ := netip.ParseAddr("192.168.1.102")

		chunk1 := AnalyzeResult{
			Ips: []HitsInfo[netip.Addr]{
				{Key: ip1, Hits: 5},
				{Key: ip2, Hits: 3},
			},
			Codes: []HitsInfo[uint16]{
				{Key: 200, Hits: 6},
				{Key: 404, Hits: 2},
			},
			TotalRequests:    8,
			UniqueUserAgents: 2,
			UserAgents: map[string]bool{
				"Mozilla/5.0": true,
				"Chrome/5.0":  true,
			},
			TimeRange: TimeRange{
				Start: time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 12, 25, 11, 0, 0, 0, time.UTC),
			},
			ProcessingStats: ProcessingStats{
				ParseErrors: 1,
			},
		}

		chunk2 := AnalyzeResult{
			Ips: []HitsInfo[netip.Addr]{
				{Key: ip1, Hits: 3},
				{Key: ip3, Hits: 4},
			},
			Codes: []HitsInfo[uint16]{
				{Key: 200, Hits: 4},
				{Key: 500, Hits: 3},
			},
			TotalRequests:    7,
			UniqueUserAgents: 1,
			UserAgents: map[string]bool{
				"Mozilla/5.0": true,
			},
			TimeRange: TimeRange{
				Start: time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 12, 25, 13, 0, 0, 0, time.UTC),
			},
			ProcessingStats: ProcessingStats{
				ParseErrors: 2,
			},
		}

		resultChan := make(chan AnalyzeResult, 2)
		resultChan <- chunk1
		resultChan <- chunk2
		close(resultChan)

		params := MergeParams{
			ChunksCount:  2,
			TopN:         10,
			Desc:         true,
			FileSize:     1000,
			WorkersCount: 2,
		}

		result := mergeResults(resultChan, params)

		assert.Equal(t, uint64(15), result.TotalRequests)
		assert.Equal(t, uint64(3), result.UniqueIPs)
		assert.Equal(t, uint64(2), result.UniqueUserAgents)
		assert.Equal(t, uint64(3), result.ProcessingStats.ParseErrors)
		assert.Equal(t, int64(1000), result.ProcessingStats.FileSize)

		// Check time range
		expectedStart := time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC)
		expectedEnd := time.Date(2023, 12, 25, 13, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedStart, result.TimeRange.Start)
		assert.Equal(t, expectedEnd, result.TimeRange.End)
	})
}
