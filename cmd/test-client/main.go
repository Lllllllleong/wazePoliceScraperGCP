package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
)

func main() {
	// As a senior developer, I want to see configurable, clear, and useful command-line flags.
	baseURL := flag.String("base-url", "http://localhost:8080", "The base URL of the alerts-service.")
	datesStr := flag.String("dates", time.Now().Format("2006-01-02"), "Comma-separated list of dates to query in YYYY-MM-DD format.")
	flag.Parse()

	if *baseURL == "" || *datesStr == "" {
		log.Fatal("Both -base-url and -dates flags are required.")
	}

	// Construct the full URL. Let's not assume the endpoint path is static.
	endpoint := "/police_alerts"
	url := fmt.Sprintf("%s%s?dates=%s", *baseURL, endpoint, *datesStr)

	fmt.Printf("--- Performance Test Starting ---\\n")
	fmt.Printf("Target URL: %s\\n", url)

	client := &http.Client{
		Timeout: 30 * time.Second, // A timeout is crucial. Never make a request without one.
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logFatalf("Failed to create request: %v", err)
	}

	startTime := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		logFatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logFatalf("Received non-OK status code: %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Now, let's process the stream. This is the critical part.
	// We need to measure if we're actually getting a stream of valid JSON.
	reader := bufio.NewReader(resp.Body)
	alertCount := 0
	errorCount := 0
	var firstByteTime, lastByteTime time.Time

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			if alertCount == 0 {
				firstByteTime = time.Now()
			}

			var alert models.PoliceAlert
			if jsonErr := json.Unmarshal(line, &alert); jsonErr != nil {
				log.Printf("ERROR: Failed to unmarshal JSON line: %v. Line: %s", jsonErr, strings.TrimSpace(string(line)))
				errorCount++
			} else {
				// For debugging, we could print the alert, but for a perf test, we just count.
				// fmt.Printf("Received alert: %s\n", alert.UUID)
				alertCount++
			}
		}

		if err == io.EOF {
			lastByteTime = time.Now()
			break
		}
		if err != nil {
			logFatalf("Error reading response body stream: %v", err)
		}
	}

	totalDuration := time.Since(startTime)
	timeToFirstByte := firstByteTime.Sub(startTime)
	streamingDuration := lastByteTime.Sub(firstByteTime)

	fmt.Printf("\\n--- Performance Test Results ---\\n")
	fmt.Printf("Total Alerts Received: %d\\n", alertCount)
	fmt.Printf("Invalid JSON Lines: %d\\n", errorCount)
	fmt.Printf("Total Request/Response Time: %v\\n", totalDuration)
	fmt.Printf("Time to First Byte (TTFB): %v\\n", timeToFirstByte)
	fmt.Printf("Streaming Duration (First to Last Byte): %v\\n", streamingDuration)

	if alertCount > 0 {
		fmt.Printf("Average Time per Alert (Streaming): %v\\n", streamingDuration/time.Duration(alertCount))
	}

	fmt.Printf("\\n--- Test Finished ---\\n")

	if errorCount > 0 {
		log.Fatalf("Test failed with %d invalid JSON lines.", errorCount)
	}
}

func logFatalf(format string, v ...interface{}) {
	log.Printf("\033[31m"+format+"\033[0m\n", v...)
	panic(fmt.Sprintf(format, v...))
}