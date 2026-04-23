package subdomain

import (
	"bufio" // for reading output
	"fmt"
	"os/exec" // for executing external program
	"strings"
)

// we split commands will be executed because of that go doesn't use shell

func RunSubfinder(domain string) []string {
	cmd := exec.Command("subfinder", "-d", domain, "-silent")

	// catch the output
	output, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Subfinder error:", err)
		return nil
	}

	// start subfinder
	err = cmd.Start()
	if err != nil {
		fmt.Println("Subfinder start error:", err)
		return nil
	}

	// create scanner
	var results []string
	scanner := bufio.NewScanner(output)

	// add founded subdomains to array
	for scanner.Scan() {
		results = append(results, scanner.Text())
	}

	cmd.Wait() // wait until command stops
	return results
}

func RunAssetfinder(domain string) []string {
	cmd := exec.Command("assetfinder", "--subs-only", domain)

	// catch output
	output, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Assetfinder error:", err)
		return nil
	}

	// start assetfinder
	err = cmd.Start()
	if err != nil {
		fmt.Println("Assetfinder start error:", err)
		return nil
	}

	// create scanner
	var results []string
	scanner := bufio.NewScanner(output)

	// add founded subdomains to array
	for scanner.Scan() {
		results = append(results, scanner.Text())
	}

	// wait until command stops
	cmd.Wait()
	return results
}

func RemoveDuplicates(input []string) []string {
	// create map
	seen := make(map[string]bool)
	var result []string

	// go through all list
	for _, sub := range input {
		if !seen[sub] {
			seen[sub] = true
			result = append(result, sub)
		}
	}
	return result
}

func GetAllSubdomains(domain string) []string {
	// channel
	resultsChan := make(chan []string)

	// start goroutine
	// subfinder
	go func() {
		fmt.Println("[*] Running subfinder...")
		res := RunSubfinder(domain)
		fmt.Println("[+] Subfinder finished:", len(res))
		resultsChan <- res
	}()

	// assetfinder
	go func() {
		fmt.Println("[*] Running assetfinder...")
		res := RunAssetfinder(domain)
		fmt.Println("[+] Assetfinder finished:", len(res))
		resultsChan <- res
	}()

	var all []string

	// 3 tools - 3 results
	for i := 0; i < 2; i++ {
		result := <-resultsChan
		all = append(all, result...)
	}

	// dedup
	all = RemoveDuplicates(all)

	// filter only target domain
	var filtered []string

	for _, sub := range all {
		if sub == "" {
			continue
		}

		if sub == domain || strings.HasSuffix(sub, "."+domain) {
			filtered = append(filtered, sub)
		}
	}

	return filtered
}
