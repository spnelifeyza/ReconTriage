package host

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type HostResult struct {
	Domain string
	URL    string
	Status int
	State  string // alive, timeout, error
	ErrMsg string
}

func CheckAlive(subdomains []string) []HostResult {
	// for workers
	jobs := make(chan string)
	// worker results
	resultsChan := make(chan HostResult)

	// concurrent request count
	workerCount := 30

	// parallel working
	for i := 0; i < workerCount; i++ {
		go func() {
			// client for each worker
			client := http.Client{
				Timeout: 3 * time.Second,
			}

			for sub := range jobs {
				// trying HTTPS first
				url := "https://" + sub
				resp, err := client.Get(url)

				// check error
				if err != nil {
					state := "error"

					// check timeout
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						state = "timeout"
					}

					// HTTP fallback
					url = "http://" + sub
					resp, err = client.Get(url)

					if err != nil {
						state = "error"

						if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
							state = "timeout"
						}

						// both fail
						resultsChan <- HostResult{
							Domain: sub,
							URL:    url,
							Status: 0,
							State:  state,
							ErrMsg: err.Error(),
						}
						continue
					}
				}

				state := "alive"
				// check for http errors
				if resp.StatusCode >= 400 {
					state = "http_error"
				}
				resultsChan <- HostResult{
					Domain: sub,
					URL:    url,
					Status: resp.StatusCode,
					State:  state,
				}

				// for preventing memory leak
				if resp != nil {
					resp.Body.Close()
				}
			}
		}()
	}

	// finish jobs
	go func() {
		for _, sub := range subdomains {
			jobs <- sub
		}
		close(jobs)
	}()

	var results []HostResult

	// keeping results
	for i := 0; i < len(subdomains); i++ {
		res := <-resultsChan
		results = append(results, res)
	}

	return results
}

func SplitResults(results []HostResult) (alive, timeout []string, errorsMap map[string][]string) {

	errorsMap = map[string][]string{
		"DNS":        {},
		"CONNECTION": {},
		"TLS":        {},
		"HTTP":       {},
		"OTHER":      {},
	}

	for _, r := range results {
		switch r.State {
		case "alive":
			alive = append(alive, r.URL)

		case "timeout":
			timeout = append(timeout, r.URL)

		case "error":
			category := ClassifyError(r.ErrMsg)
			line := r.Domain + " | " + r.ErrMsg
			errorsMap[category] = append(errorsMap[category], line)

		case "http_error":
			line := fmt.Sprintf("%s | %d", r.URL, r.Status)
			errorsMap["HTTP"] = append(errorsMap["HTTP"], line)
		}
	}
	return alive, timeout, errorsMap
}

func ClassifyError(err string) string {
	err = strings.ToLower(err)

	// DNS
	if strings.Contains(err, "no such host") {
		return "DNS"
	}

	// CONNECTION
	if strings.Contains(err, "connection refused") ||
		strings.Contains(err, "dial tcp") ||
		strings.Contains(err, "network is unreachable") {
		return "CONNECTION"
	}

	// TLS
	if strings.Contains(err, "tls") ||
		strings.Contains(err, "handshake") {
		return "TLS"
	}

	if strings.Contains(err, "http") {
		return "HTTP"
	}

	return "OTHER"
}
