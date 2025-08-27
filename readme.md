# Go nginx access log file analyzer

An example of go cli utility for analyzing nginx access files.

## Usage

```bash
go run . access.log # run analyze
go run . --gen # to generate access.log
```

## Code

Key parts of the analyzer:

### Generic data structure
```go
type HitsInfo[T comparable] struct {
  Hits int
  Key  T
}

type AnalyzeResult struct {
  Ips   []*HitsInfo[string]
  Codes []*HitsInfo[int]
  Dates []*HitsInfo[time.Time]
}
```

### Stream processing
```go
func Analyze(fpath string) (*AnalyzeResult, error) {
  file, err := os.Open(fpath)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  scanner := bufio.NewScanner(file)
  
  ipHits := make(map[string]int, 1000)
  httpCodeHits := make(map[int]int, 100)
  timeStampHits := make(map[time.Time]int, 1000)
  
  for scanner.Scan() {
    line := scanner.Text()
    log, err := parseLogEntry(line)
    if err != nil {
      fmt.Fprintln(os.Stderr, err.Error())
      continue
    }
    
    ipHits[log.ip]++
    httpCodeHits[log.statusCode]++
    timeStampHits[log.date]++
  }
  
  return createResult(ipHits, httpCodeHits, timeStampHits), nil
}
```

### Memory optimization
```go
func createResult(ipHits map[string]int, httpCodeHits map[int]int, timeStampHits map[time.Time]int) *AnalyzeResult {
  ips := make([]*HitsInfo[string], 0, len(ipHits))
  codes := make([]*HitsInfo[int], 0, len(httpCodeHits))
  dates := make([]*HitsInfo[time.Time], 0, len(timeStampHits))
  for k, v := range ipHits {
    ips = append(ips, &HitsInfo[string]{
      Hits: v,
      Key:  k,
    })
  }
  
  slices.SortFunc(ips, func(a, b *HitsInfo[string]) int {
    return b.Hits - a.Hits
  })
  
  return &AnalyzeResult{
    Ips:   ips,
    Codes: codes,
    Dates: dates,
  }
}
```

### Custom iterator
```go
type WordIter struct {
  words []string
}

func (iter *WordIter) NextOrEmpty() string {
  if len(iter.words) == 0 {
    return ""
  }
  word := iter.words[0]
  iter.words = iter.words[1:]
  return word
}
```
