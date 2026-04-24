package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"recontriage/internal/analyzer"
	"recontriage/internal/host"
	"recontriage/internal/subdomain"
	"recontriage/internal/utils"
	"sort"
	"time"
)

// color codes
const (
	Blue   = "\033[34m" // info
	Green  = "\033[32m" // success
	Yellow = "\033[33m" // warning
	Red    = "\033[31m" // error
	Reset  = "\033[0m"  // reset color
)

type FinalReport struct {
	Alive   []string                  `json:"alive"`
	Timeout []string                  `json:"timeout"`
	Errors  map[string][]string       `json:"errors"`
	Results []analyzer.AnalysisResult `json:"analysis"`
}

func main() {

	fmt.Println(Blue + `

                 /\ 
                /  \        
               /----\     
              /      \    
             /  /\    \         
            /  /  \    \    
           /__/____\____\     
              \    /
               \  /
                \/    

██████╗ ███████╗ ██████╗ ██████╗ ███╗   ██╗████████╗██████╗ ██╗ █████╗  ██████╗ ███████╗
██╔══██╗██╔════╝██╔════╝██╔═══██╗████╗  ██║╚══██╔══╝██╔══██╗██║██╔══██╗██╔════╝ ██╔════╝
██████╔╝█████╗  ██║     ██║   ██║██╔██╗ ██║   ██║   ██████╔╝██║███████║██║  ███╗█████╗  
██╔══██╗██╔══╝  ██║     ██║   ██║██║╚██╗██║   ██║   ██╔══██╗██║██╔══██║██║   ██║██╔══╝  
██║  ██║███████╗╚██████╗╚██████╔╝██║ ╚████║   ██║   ██║  ██║██║██║  ██║╚██████╔╝███████╗
╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝

        [ RECONNOITER • ANALYZE • PRIORITIZE ]

` + Reset)

	fmt.Print(Green + "=== Welcome to ReconTriage ===" + Reset + "\n\n" + "Please do your selection!\n")
	// Main menu loop
	for {
		fmt.Println(
			Green + "T" + Reset + " → Enter target\n" +
				Blue + "H" + Reset + " → How the tool works\n" +
				Red + "E" + Reset + " → Exit",
		)
		fmt.Print("Select option: ")

		var choice string
		fmt.Scan(&choice)

		switch choice {

		// TARGET OPTION
		case "T", "t":

			var domain string

			// Loop until valid domain is entered
			for {
				fmt.Print("\nEnter target > ")
				fmt.Scan(&domain)

				domain = utils.NormalizeDomain(domain)

				if !utils.IsValidDomain(domain) {
					fmt.Println(Red + "[!] Invalid domain, try again." + Reset)
					continue
				}

				break
			}

			// Start recon process
			fmt.Println(Green + "\n[✓] Target accepted. Starting ReconTriage..." + Reset)
			loading("Initializing")

			StartRecon(domain)
			return

		// HELP OPTION
		case "H", "h":

			fmt.Println(Blue + "\n[i] Showing help information..." + Reset)

			fmt.Println(Blue + `
[HELP]

ReconTriage is a recon automation tool.

It performs:
- Subdomain discovery
- Alive host detection
- Basic web analysis (title + keyword scanning)

How results are generated:
- Targets are scanned via HTTP requests
- Page content is analyzed using keyword matching
- Each target is assigned a severity score (LOW / MED / HIGH)

Output:
- all.txt      → all discovered subdomains
- alive.txt    → reachable (alive) targets
- timeout.txt  → targets that did not respond in time
- errors.txt   → targets that returned connection errors
- report.txt   → human-readable analysis report
- results.json → structured JSON output (for automation)

Notes:
- Timeout means the host did not respond within the allowed time
- Errors usually indicate connection issues, DNS problems, or refused requests
` + Reset)

		// EXIT OPTION
		case "E", "e":

			fmt.Println(Red + "[!] Exiting ReconTriage...\n" + Reset)
			return

		// INVALID INPUT
		default:

			fmt.Println(Red + "\nInvalid option." + Reset)
			fmt.Scanln()
		}
	}
}

func loading(text string) {
	fmt.Print(Yellow + "[*] " + text)

	for i := 0; i < 3; i++ {
		fmt.Print(".")
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Println(Reset)
}

func StartRecon(domain string) {
	// normalize domain
	domain = utils.NormalizeDomain(domain)

	// validate
	if !utils.IsValidDomain(domain) {
		fmt.Println(Red + "Invalid domain format!" + Reset)
		return
	}

	// print received domain
	fmt.Println(Yellow+"[*] Target:"+Reset, domain)

	// execute subdomain tools
	subs := subdomain.GetAllSubdomains(domain)
	fmt.Println(Green + "\n[+] Subdomain discovery completed." + Reset)
	fmt.Printf(Green+"[+] Total unique subdomains: %d"+Reset+"\n", len(subs))

	// alive check
	fmt.Println(Yellow + "\n[!] Checking alive hosts..." + Reset)
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

	fmt.Printf(
		Green+"[+] Alive: %d"+Reset+" | "+
			Yellow+"Timeout: %d"+Reset+" | "+
			Red+"Error: %d"+Reset+"\n",
		len(alive), len(timeout), errorCount,
	)

	// start to analyze
	RunAnalysis(alive, timeout, errorsMap, basePath)

}

func RunAnalysis(alive []string, timeout []string, errorsMap map[string][]string, basePath string) {

	fmt.Println("\n" + Blue + "[*] Running analysis..." + Reset)

	results := analyzer.Analyze(alive) // take results

	// sort by score high to low
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// write report into a file
	utils.WriteReport(filepath.Join(basePath, "report.txt"), results)

	// save results as JSON file
	report := FinalReport{
		Alive:   alive,
		Timeout: timeout,
		Errors:  errorsMap,
		Results: results,
	}

	WriteJSONReport(filepath.Join(basePath, "results.json"), report)

	fmt.Println(Green + "[+] Analysis completed." + Reset)
}

func WriteJSONReport(path string, data interface{}) {

	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating JSON file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(data)
	if err != nil {
		fmt.Println("Error writing JSON:", err)
	}
}
