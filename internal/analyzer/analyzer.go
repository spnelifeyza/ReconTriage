package analyzer

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

var titleRegex = regexp.MustCompile(`(?i)<title>(.*?)</title>`)

type AnalysisResult struct {
	URL      string
	Title    string
	Severity string
	Score    int
}

// mapping words
type KeywordConfig struct {
	High     []string `json:"high"`
	Med      []string `json:"med"`
	Low      []string `json:"low"`
	Negative []string `json:"negative"`
}

func LoadKeywords(path string) KeywordConfig {
	file, err := os.Open(path) // open file
	if err != nil {
		fmt.Println("Error opening keywords file:", err)
		return KeywordConfig{}
	}
	defer file.Close()

	var config KeywordConfig
	err = json.NewDecoder(file).Decode(&config) // decode json file
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return KeywordConfig{}
	}
	// return words
	return config
}

func FetchContent(url string) (string, string) {
	// HTTP client with timeout
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	maxRetries := 3 // max retry attempts

	for i := 0; i < maxRetries; i++ {

		resp, err := client.Get(url)
		if err != nil {
			// retry only for network-related errors
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// stop retry if page does not exist
		if resp.StatusCode == 404 {
			resp.Body.Close()
			return "", ""
		}

		// handle rate limiting (retry after delay)
		if resp.StatusCode == 429 {
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		// 403 = restricted but interesting, do not retry
		if resp.StatusCode == 403 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			body := string(bodyBytes)

			match := titleRegex.FindStringSubmatch(body)
			title := ""
			if len(match) > 1 {
				title = match[1]
			}

			return title, body
		}

		// normal response handling
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			// retry if body read fails
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		body := string(bodyBytes)

		// limit body size to prevent memory issues
		if len(body) > 200000 {
			body = body[:200000]
		}

		// extract title from HTML
		match := titleRegex.FindStringSubmatch(body)
		title := ""
		if len(match) > 1 {
			title = match[1]
		}

		return title, body
	}

	// return empty if all retries fail
	return "", ""
}

func Analyze(alive []string) []AnalysisResult {
	cfg := LoadKeywords("./configs/keywords.json")
	// create jobs
	jobs := make(chan string)
	resultsChan := make(chan AnalysisResult)

	workerCount := 10 * runtime.NumCPU()

	// workers
	for i := 0; i < workerCount; i++ {
		go func() {
			for url := range jobs {
				// create correct urls
				if !strings.HasPrefix(url, "http") {
					url = "http://" + url
				}
				// getting slow down randomly
				time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

				if IsStatic(url) {
					resultsChan <- AnalysisResult{
						URL:      url,
						Title:    "",
						Severity: "LOW",
						Score:    0,
					}
					continue
				}

				title, body := FetchContent(url)

				score := CalculateScore(url, title, body, cfg)
				score += CheckCriticalPaths(url)

				severity := GetSeverity(score)

				resultsChan <- AnalysisResult{
					URL:      url,
					Title:    title,
					Severity: severity,
					Score:    score,
				}
			}
		}()
	}

	// send jobs
	go func() {
		for _, url := range alive {
			jobs <- url
		}
		close(jobs)
	}()

	var results []AnalysisResult

	// collect results
	for i := 0; i < len(alive); i++ {
		res := <-resultsChan
		results = append(results, res)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}

func CheckCriticalPaths(url string) int {
	url = strings.ToLower(url)

	if strings.Contains(url, "admin") ||
		strings.Contains(url, "login") ||
		strings.Contains(url, "dashboard") {
		return 5
	}
	return 0
}

func CalculateScore(url, title, body string, cfg KeywordConfig) int {
	score := 0

	content := strings.ToLower(url + " " + title)
	bodyLower := strings.ToLower(body)

	// HIGH keywords
	for _, k := range cfg.High {
		if strings.Contains(content, strings.ToLower(k)) {
			score += 5
		}
	}

	// MED keywords
	for _, k := range cfg.Med {
		if strings.Contains(content, strings.ToLower(k)) {
			score += 3
		}
	}

	// LOW keywords
	for _, k := range cfg.Low {
		if strings.Contains(content, strings.ToLower(k)) {
			score += 1
		}
	}

	// NEGATIVE keywords
	for _, k := range cfg.Negative {
		if strings.Contains(bodyLower, strings.ToLower(k)) {
			score -= 3
		}
	}

	return score
}

func GetSeverity(score int) string {
	if score >= 8 {
		return "HIGH"
	} else if score >= 4 {
		return "MEDIUM"
	}
	return "LOW"
}

func IsStatic(url string) bool {
	url = strings.ToLower(url)

	staticExt := []string{
		".css", ".js", ".png", ".jpg", ".jpeg",
		".gif", ".svg", ".ico", ".woff", ".ttf",
	}

	// check if it a static file
	for _, ext := range staticExt {
		if strings.HasSuffix(url, ext) {
			return true
		}
	}
	return false
}
