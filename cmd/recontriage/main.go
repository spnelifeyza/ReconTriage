package main

import (
	"fmt"
	"os"
	"path/filepath"
	"recontriage/internal/analyzer"
	"recontriage/internal/host"
	"recontriage/internal/subdomain"
	"recontriage/internal/utils"
)

func main() {
	// check if user provided a domain argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <domain>")
		return
	}

	// get domain from command line
	domain := os.Args[1]
	StartRecon(domain)

}

func StartRecon(domain string) {
	// normalize domain
	domain = utils.NormalizeDomain(domain)

	// validate
	if !utils.IsValidDomain(domain) {
		fmt.Println("Invalid domain format!")
		return
	}

	// print received domain
	fmt.Println("[*] Target:", domain)

	// execute subdomain tools
	subs := subdomain.GetAllSubdomains(domain)
	fmt.Println("\n[+] Subdomain discovery completed.")
	fmt.Println("[+] Total unique subdomains:", len(subs))

	// alive check
	fmt.Println("\n[*] Checking alive hosts...")
	results := host.CheckAlive(subs)

	// split results into categories
	alive, timeout, errorsMap := host.SplitResults(results)

	// create output folder per domain
	basePath := "outputs/" + domain + "/subdomains/"
	os.MkdirAll(basePath, os.ModePerm)

	// write to file
	utils.WriteToFile(filepath.Join(basePath, "all.txt"), subs)
	utils.WriteToFile(filepath.Join(basePath, "alive.txt"), alive)
	utils.WriteToFile(filepath.Join(basePath, "timeout.txt"), timeout)
	utils.WriteErrorsToFile(filepath.Join(basePath, "errors.txt"), errorsMap)

	// count errors
	errorCount := 0
	for _, list := range errorsMap {
		errorCount += len(list)
	}

	fmt.Printf("[+] Alive: %d | Timeout: %d | Error: %d\n", len(alive), len(timeout), errorCount)

	// start to analyze
	RunAnalysis(alive, basePath)

}

func RunAnalysis(alive []string, basePath string) {

	fmt.Println("\n[*] Running analysis...")

	results := analyzer.Analyze(alive) // take results

	// write report into a file
	utils.WriteReport(filepath.Join(basePath, "report.txt"), results)
	fmt.Println("[+] Analysis completed.")
}
