package generator

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type LogEntry struct {
	IP        string
	Time      string
	Method    string
	Path      string
	Protocol  string
	Status    int
	Size      int
	Referer   string
	UserAgent string
}

func GenerateAccessLog() {
	rand.Seed(time.Now().UnixNano())

	// Config
	const outputFile = "access.log"
	const numEntries = 10000 // Rows to generate

	// Creating a file
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Create file error: %v\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Generating logs
	for i := 0; i < numEntries; i++ {
		entry := generateLogEntry()
		logLine := formatLogEntry(entry)
		writer.WriteString(logLine + "\n")

		// Progress
		if i%1000 == 0 {
			fmt.Printf("Generated %d records of %d\n", i, numEntries)
		}
	}

	fmt.Printf("Successfully created a file %s with %d records\n", outputFile, numEntries)
}

func generateLogEntry() LogEntry {
	// Values with different probability
	ips := []string{
		"192.168.1.100", "192.168.1.101", "192.168.1.102", // Often
		"10.0.0.50", "10.0.0.51", // Medium
		"172.16.0.10", "172.16.0.11", "172.16.0.12", // Rare
	}

	paths := []string{
		"/", "/", "/", "/", // Very often
		"/about", "/about", // Often
		"/contact",               // Medium
		"/products", "/products", // Often
		"/products/electronics",          // Medium
		"/products/books",                // Medium
		"/products/clothing",             // Medium
		"/admin",                         // Rare
		"/api/v1/users", "/api/v1/users", // Often API
		"/api/v1/orders",     // Medium API
		"/api/v1/products",   // Medium API
		"/user/profile",      // Rare
		"/search", "/search", // Often
		"/cart", "/cart", // Often
		"/checkout",      // Medium
		"/blog/post/123", // Rare
		"/blog/post/456",
		"/blog/post/789",
	}

	statuses := []int{200, 200, 200, 200, 200, 304, 404, 500} // 200 - The most often

	// User-Agents
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
		"Googlebot/2.1 (+http://www.google.com/bot.html)",
		"curl/7.68.0",
	}

	referrers := []string{
		"-",
		"https://www.google.com/",
		"https://www.bing.com/",
		"https://example.com/",
		"https://socialmedia.com/post/123",
	}

	now := time.Now()
	randomOffset := time.Duration(rand.Intn(86400)) * time.Second // Random time (24h)
	logTime := now.Add(-randomOffset)

	return LogEntry{
		IP:        weightedRandom(ips, []int{40, 35, 25, 20, 20, 10, 8, 5}),
		Time:      logTime.Format("[02/Jan/2006:15:04:05 -0700]"),
		Method:    weightedRandom([]string{"GET", "POST", "PUT", "DELETE"}, []int{80, 15, 3, 2}),
		Path:      weightedRandom(paths, generateWeights(len(paths))),
		Protocol:  "HTTP/1.1",
		Status:    statuses[rand.Intn(len(statuses))],
		Size:      rand.Intn(5000) + 200, // File size 200-5200 bytes
		Referer:   referrers[rand.Intn(len(referrers))],
		UserAgent: userAgents[rand.Intn(len(userAgents))],
	}
}

// returns random element with weights in mind
func weightedRandom(items []string, weights []int) string {
	if len(items) != len(weights) {
		return items[rand.Intn(len(items))]
	}

	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	r := rand.Intn(totalWeight)
	for i, w := range weights {
		r -= w
		if r <= 0 {
			return items[i]
		}
	}

	return items[len(items)-1]
}

// (first elements gain more weight)
func generateWeights(n int) []int {
	weights := make([]int, n)
	for i := range weights {
		weights[i] = n - i // Losing weights
	}
	return weights
}

func formatLogEntry(entry LogEntry) string {
	return fmt.Sprintf("%s - - %s \"%s %s %s\" %d %d \"%s\" \"%s\"",
		entry.IP,
		entry.Time,
		entry.Method,
		entry.Path,
		entry.Protocol,
		entry.Status,
		entry.Size,
		entry.Referer,
		entry.UserAgent,
	)
}
