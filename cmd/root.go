package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kostayne/go-nginx-analyzer/analyzer"
	"github.com/spf13/cobra"
)

type Flags struct {
	FilePath string
	Top      int
	IsDesc   bool
	DatesBy  string
	Output   string
}

var rootCmd = &cobra.Command{
	Use:   "nginx-an <path-to-access.log>",
	Short: "Nginx access log analyzer",
	Long:  "Nginx access log analyzer written in go.",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		flags, err := parseFlags(cmd, args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		res, err := analyzer.Analyze(flags.FilePath, flags.Top, flags.IsDesc, flags.DatesBy)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		printSummary(res)
		printTopInfo(res.Ips, "Top ips", flags.Top)
		printTopInfo(res.Codes, "Top status codes", flags.Top)
		printTopInfo(res.Dates, "Top dates", flags.Top)
		printProcessingStats(res.ProcessingStats)

		if flags.Output != "" {
			err = saveToFile(*res, flags.Output)
			if err != nil {
				fmt.Println("Error saving to file:", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.PersistentFlags().Bool("desc", true, "sort in descending order")
	rootCmd.PersistentFlags().Bool("asc", false, "sort in ascending order")
	rootCmd.PersistentFlags().Int("top", 10, "limit the number of results")
	rootCmd.PersistentFlags().String("dates-by", "none", "group dates by: none, hour, day")
	rootCmd.PersistentFlags().StringP("output", "o", "", "json output file name")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func saveToFile(res analyzer.AnalyzeResult, fileName string) error {
	jsonData, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func parseFlags(cmd *cobra.Command, args []string) (*Flags, error) {
	filePath := args[0]

	isDesc, err := parseSortFlags(cmd)
	if err != nil {
		return nil, err
	}

	top, err := parseTopFlag(cmd)
	if err != nil {
		return nil, err
	}

	datesBy, err := parseDatesByFlag(cmd)
	if err != nil {
		return nil, err
	}

	output, err := parseOutputFlag(cmd)
	if err != nil {
		return nil, err
	}

	return &Flags{
		FilePath: filePath,
		Top:      top,
		IsDesc:   isDesc,
		DatesBy:  datesBy,
		Output:   output,
	}, nil
}

func parseSortFlags(cmd *cobra.Command) (bool, error) {
	desc, descErr := cmd.PersistentFlags().GetBool("desc")
	if descErr != nil {
		return false, fmt.Errorf("failed to get desc flag: %w", descErr)
	}

	asc, ascErr := cmd.PersistentFlags().GetBool("asc")
	if ascErr != nil {
		return false, fmt.Errorf("failed to get asc flag: %w", ascErr)
	}

	if desc && asc {
		return false, fmt.Errorf("cannot use both --desc and --asc flags")
	}

	return !asc, nil
}

func parseTopFlag(cmd *cobra.Command) (int, error) {
	top, topErr := cmd.PersistentFlags().GetInt("top")
	if topErr != nil {
		return 0, fmt.Errorf("failed to get top flag: %w", topErr)
	}

	if top <= 0 {
		return 0, fmt.Errorf("top must be greater than 0")
	}

	return top, nil
}

func parseDatesByFlag(cmd *cobra.Command) (string, error) {
	datesBy, datesByErr := cmd.PersistentFlags().GetString("dates-by")
	if datesByErr != nil {
		return "", fmt.Errorf("failed to get dates-by flag: %w", datesByErr)
	}

	if !isValidDatesByOption(datesBy) {
		return "", fmt.Errorf("dates-by must be one of: none, hour, day")
	}

	return datesBy, nil
}

func parseOutputFlag(cmd *cobra.Command) (string, error) {
	output, outputErr := cmd.PersistentFlags().GetString("output")
	if outputErr != nil {
		return "", fmt.Errorf("failed to get output flag: %w", outputErr)
	}

	return output, nil
}

func isValidDatesByOption(datesBy string) bool {
	validOptions := []string{"none", "hour", "day"}
	for _, option := range validOptions {
		if datesBy == option {
			return true
		}
	}
	return false
}

func printSummary(res *analyzer.AnalyzeResult) {
	fmt.Println("SUMMARY")
	fmt.Println(strings.Repeat("=", 7))
	fmt.Printf("Total Requests: %d\n", res.TotalRequests)
	fmt.Printf("Unique IPs: %d\n", res.UniqueIPs)
	fmt.Printf("Unique User Agents: %d\n", res.UniqueUserAgents)
	fmt.Printf("Time Range: %s to %s\n", res.TimeRange.Start.Format("2006-01-02 15:04:05"), res.TimeRange.End.Format("2006-01-02 15:04:05"))
	fmt.Println()
}

func printTopInfo[T comparable](hitsInfo []analyzer.HitsInfo[T], msg string, limit int) {
	fmt.Println(msg)
	fmt.Println(strings.Repeat("=", len(msg)))

	for i, info := range hitsInfo[:min(limit, len(hitsInfo))] {
		// Special formatting for time.Time
		if timeVal, ok := any(info.Key).(time.Time); ok {
			fmt.Printf("%d %s: %d \n", i+1, timeVal.Format("2006-01-02 15:04:05"), info.Hits)
		} else {
			fmt.Printf("%d %v: %d \n", i+1, info.Key, info.Hits)
		}
	}
	fmt.Println()
}

func printProcessingStats(stats struct {
	FileSize    int64  `json:"fileSize"`
	ParseErrors uint64 `json:"parseErrors"`
}) {
	fmt.Println("PROCESSING STATISTICS")
	fmt.Println(strings.Repeat("=", 22))
	fmt.Printf("File Size: %.2f MB\n", float64(stats.FileSize)/(1024*1024))
	fmt.Printf("Parse Errors: %d\n", stats.ParseErrors)
	fmt.Println()
}
