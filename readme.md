# Go nginx access log file analyzer

An example of go cli utility for analyzing nginx access files.

## Usage

```bash
go run . access.log # run analyzer
go run . --gen # to generate access.log
```

## Code

Key parts of the high-performance analyzer:

### Generic data structures with type safety
```go
type HitsInfo[T comparable] struct {
	Hits uint64 `json:"hits"`
	Key  T      `json:"key"`
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

	// Time range
	TimeRange TimeRange `json:"timeRange"`
}
```

### Multi-threaded chunk processing with memory mapping
```go
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

	// Start worker goroutines
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

	// Send chunks to workers
	for _, chunk := range chunks {
		chunkChan <- chunk
	}
	close(chunkChan)

	wg.Wait()
	close(resultChan)

	// Merge results from all chunks
	res := mergeResults(resultChan, MergeParams{
		ChunksCount:  int(chunksCount),
		TopN:         topN,
		Desc:         desc,
		FileSize:     fileSize,
		WorkersCount: workersCount,
	})
	return &res, nil
}
```

### Generic sorting and aggregation
```go
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
```
