package utils

import (
	"fmt"
	"os"
	"recontriage/internal/analyzer"
	"regexp"
	"strings"
)

// normalize domain
func NormalizeDomain(input string) string {
	domain := strings.TrimSpace(input)

	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimSuffix(domain, "/")

	return domain
}

// validate domain using regex
func IsValidDomain(domain string) bool {
	// basic domain regex
	regex := `^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)

	return re.MatchString(domain)
}

// write subdomains into a file
func WriteToFile(filename string, data []string) {
	// create file
	file, err := os.Create(filename)
	if err != nil {
		return // leave if there is error
	}

	// close file
	defer file.Close()

	// write all data into file
	for _, line := range data {
		file.WriteString(line + "\n")
	}
}

func WriteErrorsToFile(filename string, errorsMap map[string][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close() // close the file

	order := []string{"DNS", "CONNECTION", "TLS", "HTTP", "OTHER"}

	// total error count
	total := 0
	for _, list := range errorsMap {
		total += len(list)
	}

	// header
	file.WriteString("# ReconTriage Error Report\n")
	file.WriteString(fmt.Sprintf("# Total Errors: %d\n\n", total))

	for _, category := range order {
		list := errorsMap[category]

		// jump if a category is empty
		if len(list) == 0 {
			continue
		}

		// write categories
		file.WriteString("# ================== " + category + " ERRORS ==================\n")

		for _, line := range list {
			file.WriteString(line + "\n")
		}

		file.WriteString("\n")
	}

	return nil
}

func WriteReport(filename string, results []analyzer.AnalysisResult) {
	file, err := os.Create(filename) // create file
	if err != nil {
		fmt.Println("Error creating report file:", err)
		return
	}
	defer file.Close()

	// write report
	for _, r := range results {
		line := fmt.Sprintf("[%s] %s → %s\n", r.Severity, r.URL, r.Title)
		file.WriteString(line)
	}
}
