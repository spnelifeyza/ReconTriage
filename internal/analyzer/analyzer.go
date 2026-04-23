package analyzer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type AnalysisResult struct {
	URL      string
	Title    string
	Severity string
}

// mapping words
type KeywordConfig struct {
	High []string `json:"high"`
	Med  []string `json:"med"`
	Low  []string `json:"low"`
}

func FetchTitle(url string) string {
	// create HTTP client
	client := http.Client{
		Timeout: 4 * time.Second,
	}

	// send GET request
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close() // prevent memory leak

	// take first 2KB - title on top
	buf := make([]byte, 2048)
	n, _ := resp.Body.Read(buf)

	// convert byte to string
	body := string(buf[:n])

	re := regexp.MustCompile(`(?i)<title>(.*?)</title>`) // create regex
	match := re.FindStringSubmatch(body)                 // find title

	if len(match) > 1 {
		return match[1]
	}

	return ""
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

func AnalyzeURL(url, title string, cfg KeywordConfig) string {

	data := strings.ToLower(url + " " + title)
	// check words for severity
	// HIGH
	for _, kw := range cfg.High {
		if strings.Contains(data, kw) {
			return fmt.Sprintf("HIGH (%s)", kw)
		}
	}

	// MED
	for _, kw := range cfg.Med {
		if strings.Contains(data, kw) {
			return fmt.Sprintf("MED (%s)", kw)
		}
	}

	// LOW
	for _, kw := range cfg.Low {
		if strings.Contains(data, kw) {
			return fmt.Sprintf("LOW (%s)", kw)
		}
	}

	return "INFO"
}

func Analyze(alive []string) []AnalysisResult {
	cfg := LoadKeywords("./configs/keywords.json")
	// create jobs
	jobs := make(chan string)
	resultsChan := make(chan AnalysisResult)

	workerCount := 20

	// workers
	for i := 0; i < workerCount; i++ {
		go func() {
			for url := range jobs {

				title := FetchTitle(url)
				severity := AnalyzeURL(url, title, cfg)

				resultsChan <- AnalysisResult{
					URL:      url,
					Title:    title,
					Severity: severity,
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

	return results
}
