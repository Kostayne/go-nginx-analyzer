package analyzer

import (
	"errors"
	"log"
	"net/netip"
	"os"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/edsrzf/mmap-go"
	"github.com/kostayne/go-nginx-analyzer/parser"
)

// Chunk size in bytes
const CHUNK_SIZE = 1024 * 1024 * 100 // 100MB
const CHUNK_OVERLAP = 1024 * 2       // 2KB
const AVG_LINE_SIZE = 400            // 400 bytes

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type ProcessingStats struct {
	FileSize    int64  `json:"fileSize"`
	ParseErrors uint64 `json:"parseErrors"`
}

type AnalyzeResult struct {
	// Top hits by category
	Ips   []HitsInfo[netip.Addr] `json:"ips"`
	Codes []HitsInfo[uint16]     `json:"codes"`
	Dates []HitsInfo[time.Time]  `json:"dates"`

	// Summary statistics
	TotalRequests    uint64 `json:"totalRequests"`
	UniqueIPs        uint64 `json:"uniqueIps"`
	UniqueUserAgents uint64 `json:"uniqueUserAgents"`

	// Status code distribution
	StatusCodes []HitsInfo[uint16] `json:"statusCodes"`

	// User agents for aggregation
	UserAgents map[string]bool `json:"userAgents"`

	// Time range
	TimeRange TimeRange `json:"timeRange"`

	// Processing statistics
	ProcessingStats ProcessingStats `json:"processingStats"`
}

type HitsInfo[T comparable] struct {
	Hits uint64 `json:"hits"`
	Key  T      `json:"key"`
}

type Chunk struct {
	fileName string
	startPos int64
	endPos   int64
}

type WorkerInfo struct {
	wg         *sync.WaitGroup
	chunkChan  <-chan Chunk
	resultChan chan<- AnalyzeResult
	fileName   string
	desc       bool
	topN       int
	groupBy    string
}

type MergeParams struct {
	ChunksCount  int
	TopN         int
	Desc         bool
	FileSize     int64
	WorkersCount int
}

type ProcessParams struct {
	TopN    int    `json:"topN"`
	Desc    bool   `json:"desc"`
	GroupBy string `json:"groupBy"`
}

func Analyze(fpath string, topN int, desc bool, datesBy string) (*AnalyzeResult, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fstat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fstat.Size()
	chunksCount := max(int(fileSize/CHUNK_SIZE), 1)
	workersCount := runtime.NumCPU()

	chunks := make([]Chunk, chunksCount)
	for i := 0; i < chunksCount; i++ {
		chunks[i] = newChunk(i, fpath, fileSize)
	}

	chunkChan := make(chan Chunk, chunksCount)
	resultChan := make(chan AnalyzeResult, chunksCount)

	wg := sync.WaitGroup{}

	for i := 0; i < workersCount; i++ {
		wg.Add(1)

		wi := &WorkerInfo{
			wg:         &wg,
			chunkChan:  chunkChan,
			resultChan: resultChan,
			fileName:   fpath,
			desc:       desc,
			topN:       topN,
			groupBy:    datesBy,
		}
		go worker(wi)
	}

	for _, chunk := range chunks {
		chunkChan <- chunk
	}
	close(chunkChan)

	wg.Wait()
	close(resultChan)

	mergeParams := MergeParams{
		ChunksCount:  int(chunksCount),
		TopN:         topN,
		Desc:         desc,
		FileSize:     fileSize,
		WorkersCount: workersCount,
	}
	res := mergeResults(resultChan, mergeParams)
	return &res, nil
}

func mergeResults(resChan <-chan AnalyzeResult, params MergeParams) AnalyzeResult {
	ipHits := make([]HitsInfo[netip.Addr], 0, params.TopN*params.ChunksCount)
	codeHits := make([]HitsInfo[uint16], 0, params.TopN*params.ChunksCount)
	dateHits := make([]HitsInfo[time.Time], 0, params.TopN*params.ChunksCount)

	// Aggregate statistics
	totalRequests := uint64(0)
	parseErrors := uint64(0)
	uniqueIPs := make(map[netip.Addr]bool)
	uniqueUserAgents := make(map[string]bool)

	var timeStart, timeEnd time.Time
	timeStartSet := false

	for res := range resChan {
		ipHits = append(ipHits, res.Ips...)
		codeHits = append(codeHits, res.Codes...)
		dateHits = append(dateHits, res.Dates...)

		// Aggregate summary statistics
		totalRequests += res.TotalRequests
		parseErrors += res.ProcessingStats.ParseErrors

		// Track unique IPs and user agents
		for _, ipInfo := range res.Ips {
			uniqueIPs[ipInfo.Key] = true
		}

		// Track unique user agents from this chunk
		for userAgent := range res.UserAgents {
			uniqueUserAgents[userAgent] = true
		}

		// Track time range
		if !timeStartSet || res.TimeRange.Start.Before(timeStart) {
			timeStart = res.TimeRange.Start
			timeStartSet = true
		}
		if res.TimeRange.End.After(timeEnd) {
			timeEnd = res.TimeRange.End
		}
	}

	result := AnalyzeResult{
		Ips:              processHitsInfo(ipHits, params.TopN, params.Desc),
		Codes:            processHitsInfo(codeHits, params.TopN, params.Desc),
		Dates:            processHitsInfo(dateHits, params.TopN, params.Desc),
		TotalRequests:    totalRequests,
		UniqueIPs:        uint64(len(uniqueIPs)),
		UniqueUserAgents: uint64(len(uniqueUserAgents)),
		TimeRange: struct {
			Start time.Time `json:"start"`
			End   time.Time `json:"end"`
		}{
			Start: timeStart,
			End:   timeEnd,
		},
		ProcessingStats: struct {
			FileSize    int64  `json:"fileSize"`
			ParseErrors uint64 `json:"parseErrors"`
		}{
			FileSize:    params.FileSize,
			ParseErrors: parseErrors,
		},
	}

	return result
}

func worker(w *WorkerInfo) {
	defer w.wg.Done()

	file, err := os.Open(w.fileName)
	if err != nil {
		log.Fatalf("Failed to open file %s: %s", w.fileName, err.Error())
		return
	}
	defer file.Close()

	for chunk := range w.chunkChan {
		processParams := ProcessParams{
			TopN:    w.topN,
			Desc:    w.desc,
			GroupBy: w.groupBy,
		}
		result, err := processChunk(chunk, file, processParams)
		if err != nil {
			log.Fatalf("error processing chunk %s: %v", chunk.fileName, err)
		}

		w.resultChan <- result
	}
}

func processChunk(chunk Chunk, file *os.File, params ProcessParams) (AnalyzeResult, error) {
	ipsMap := make(map[netip.Addr]uint64)
	codesMap := make(map[uint16]uint64)
	datesMap := make(map[time.Time]uint64)
	statusCodes := make(map[uint16]uint64)
	uniqueUserAgents := make(map[string]bool)

	totalRequests := uint64(0)
	parseErrors := uint64(0)

	var timeStart, timeEnd time.Time
	timeStartSet := false

	chunkLen := int(chunk.endPos - chunk.startPos)
	mmapData, err := mmap.MapRegion(file, chunkLen, 0, mmap.RDONLY, chunk.startPos)
	if err != nil {
		return AnalyzeResult{}, err
	}
	defer mmapData.Unmap()

	curPos := 0

	// skip first line if it's not the first chunk
	// because it's already processed in the previous chunk
	if chunk.startPos != 0 {
		curPos = findNewLineIndex(mmapData, 0)

		if curPos == -1 {
			return AnalyzeResult{}, errors.New("no new line found")
		}

		curPos++
	}

	maxPos := int(chunk.endPos - chunk.startPos)

	// process the rest of the chunk
	for curPos < maxPos {
		nextLineIndex := findNewLineIndex(mmapData, curPos)

		// if there is no new line
		// then read to the end
		if nextLineIndex == -1 {
			nextLineIndex = maxPos
		}

		curLineStr := string(mmapData[curPos:nextLineIndex])

		logEntry, err := parser.ParseLogEntry(curLineStr)
		if err != nil {
			log.Print(err.Error())
			parseErrors++
			curPos = nextLineIndex + 1
			continue
		}

		totalRequests++
		ipsMap[logEntry.Ip]++
		codesMap[logEntry.StatusCode]++

		// Group dates if needed
		dateKey := logEntry.Date
		if params.GroupBy == "hour" {
			dateKey = time.Date(logEntry.Date.Year(), logEntry.Date.Month(), logEntry.Date.Day(), logEntry.Date.Hour(), 0, 0, 0, logEntry.Date.Location())
		} else if params.GroupBy == "day" {
			dateKey = time.Date(logEntry.Date.Year(), logEntry.Date.Month(), logEntry.Date.Day(), 0, 0, 0, 0, logEntry.Date.Location())
		}
		datesMap[dateKey]++

		statusCodes[logEntry.StatusCode]++
		uniqueUserAgents[logEntry.UserAgent] = true

		// Track time range
		if !timeStartSet || logEntry.Date.Before(timeStart) {
			timeStart = logEntry.Date
			timeStartSet = true
		}
		if logEntry.Date.After(timeEnd) {
			timeEnd = logEntry.Date
		}

		curPos = nextLineIndex + 1
	}

	ipHits := getHitsInfo(ipsMap, params.TopN, params.Desc)
	codeHits := getHitsInfo(codesMap, params.TopN, params.Desc)
	dateHits := getHitsInfo(datesMap, params.TopN, params.Desc)

	res := AnalyzeResult{
		Ips:              *ipHits,
		Codes:            *codeHits,
		Dates:            *dateHits,
		TotalRequests:    totalRequests,
		UniqueUserAgents: uint64(len(uniqueUserAgents)),
		UserAgents:       uniqueUserAgents,
		TimeRange: TimeRange{
			Start: timeStart,
			End:   timeEnd,
		},
		ProcessingStats: ProcessingStats{
			ParseErrors: parseErrors,
		},
	}

	return res, nil
}

func getHitsInfo[T comparable](m map[T]uint64, topN int, desc bool) *[]HitsInfo[T] {
	hitsArr := make([]HitsInfo[T], 0, len(m))

	for k, hits := range m {
		hitsArr = append(hitsArr, HitsInfo[T]{
			Hits: hits,
			Key:  k,
		})
	}

	slices.SortFunc(hitsArr, func(a HitsInfo[T], b HitsInfo[T]) int {
		if desc {
			return getSortCompareResultAsc(b.Hits, a.Hits)
		}

		return getSortCompareResultAsc(a.Hits, b.Hits)
	})

	hitsArr = hitsArr[0:min(len(hitsArr), int(topN))]
	return &hitsArr
}

func findNewLineIndex(data []byte, start int) int {
	for i := start; i < len(data); i++ {
		if data[i] == '\n' {
			return i
		}
	}

	return -1
}

func processHitsInfo[T comparable](hits []HitsInfo[T], topN int, desc bool) []HitsInfo[T] {
	slices.SortFunc(hits, func(a HitsInfo[T], b HitsInfo[T]) int {
		if desc {
			return getSortCompareResultAsc(b.Hits, a.Hits)
		}

		return getSortCompareResultAsc(a.Hits, b.Hits)
	})

	return hits[0:min(len(hits), topN)]
}

func newChunk(index int, fileName string, fileSize int64) Chunk {
	start := int64(index*CHUNK_SIZE - CHUNK_OVERLAP)
	end := start + CHUNK_SIZE + CHUNK_OVERLAP

	if start < 0 {
		start = 0
	}

	if end >= fileSize {
		end = fileSize - 1
	}

	return Chunk{
		fileName: fileName,
		startPos: start,
		endPos:   end,
	}
}

func getSortCompareResultAsc(a, b uint64) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}

	return 0
}
