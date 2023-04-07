package cli

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/deanrtaylor1/gosearch/bm25"
	"github.com/deanrtaylor1/gosearch/tfidf"
	"github.com/deanrtaylor1/gosearch/util"
	webcrawler "github.com/deanrtaylor1/gosearch/web-crawler"
	"github.com/tebeka/snowball"
)

//CLI Interface of GoSearch

// Utility function to show the user the current status of the indexing and crawling processes
func logStatus(indexing, crawling bool, model *bm25.Model) {
	indexState := "✓"
	if indexing {
		indexState = "⌛"
	}

	crawlState := "✓"
	if crawling {
		crawlState = "⌛"
	}

	model.ModelLock.Lock()
	totalDocs := model.DocCount
	totalTerms := model.TermCount
	model.ModelLock.Unlock()

	fmt.Printf(util.TerminalGreen+"Indexing Status: %s | Crawling Status: %s | %v documents scanned | %v terms indexed\n"+util.TerminalReset, indexState, crawlState, totalDocs, totalTerms)
	fmt.Println("Type your query or press Ctrl+C to exit")
}

// Clean up the CLI response to remove the bullet point
func formatCliResponse(response string) string {
	return strings.Replace(response, "○ ", "", -1)
}

// Utility function to get a single input from the user
func getSingleInputPrompt(message string) string {
	prompt := &survey.Input{
		Message: message,
	}

	var input string
	err := survey.AskOne(prompt, &input)
	if err != nil {
		log.Fatal(err)
	}

	return input
}

// Utility function to get the website to crawl
func GetNewWebsitePrompt() string {
	return getSingleInputPrompt("Enter a website to crawl:")
}

// Start the CLI
func InitialPrompt(model *bm25.Model) {
	files, err := os.ReadDir("./indexes")
	if err != nil {
		log.Fatal(err)
	}

	directories := []string{}
	for _, f := range files {
		if f.IsDir() {
			if f.Name() == "src" || f.Name() == ".git" {
				continue
			}
			directories = append(directories, "○ "+f.Name())
		}
	}
	directories = append(directories, "○ Crawl And Index")

	prompt := &survey.Select{
		Message: "Select a directory to index:",
		Options: directories,
	}

	var selectedDirectory string
	err = survey.AskOne(prompt, &selectedDirectory)
	if err != nil {
		log.Fatal(err)
	}

	initialSelection := formatCliResponse(selectedDirectory)
	//fmt.Printf("Selected directory: %s\n", initialSelection)

	switch initialSelection {
	case "Crawl And Index":
		fmt.Println("Crawling and indexing directory: ", initialSelection)
		newSite := GetNewWebsitePrompt()
		InitCrawl(newSite, model)
	default:
		selectedIndex := formatCliResponse(selectedDirectory)
		fmt.Println("Starting server and indexing directory: ", selectedIndex)
		model.Name = selectedIndex
		go func() {
			logStatus(true, false, model)
			bm25.LoadCachedGobToModel("./indexes/"+selectedIndex, model)
			model.ModelLock.Lock()
			model.DA = float32(model.TermCount) / float32(model.DocCount)
			model.IsComplete = true
			model.ModelLock.Unlock()

			logStatus(false, false, model)

		}()
		StartQueryPrompt(model)
	}
}

// Get the query from the user to search the model
func StartQueryPrompt(model *bm25.Model) {

	prompt := &survey.Input{
		Message: "Enter a query:",
	}

	var query string
	fmt.Println()
	err := survey.AskOne(prompt, &query)
	if err != nil {
		log.Fatal(err)
	}

	startQuery(query, model)
}

// Start the query process
func startQuery(query string, model *bm25.Model) {

	start := time.Now()
	stemmer, err := snowball.New("english")
	if err != nil {
		log.Fatal(err)
	}

	defer stemmer.Close()

	query = strings.ToLower(query)
	var count int
	result, count := bm25.CalculateBm25(model, query)

	var max int
	if len(result) < 20 {
		max = len(result)
	} else {
		max = 20
	}

	// for i := 0; i < max; i++ {
	// 	log.Println(result[i].Path, " => ", result[i].TF)
	// }

	if err != nil {
		log.Println(err)
		return
	}

	if result[0].TF == 0 {
		log.Println("Query too generic, ranking with tf-idf")

		result, count = tfidf.CalculateTfidf(model, query)

		// for i := 0; i < max; i++ {
		// 	log.Println(result[i].Path, " => ", result[i].TF)
		// }

	}

	var data []bm25.ResultsMap

	if result[0].TF == 0 {
		data = []bm25.ResultsMap{{
			Path: "No results found",
			TF:   0,
		}}
	} else {
		data = bm25.FilterResults(result[:max], bm25.IsGreaterThanZero)
	}

	resultsList := []string{}
	for _, r := range data {
		resultsList = append(resultsList, "○ "+r.Name)
	}
	resultsList = append(resultsList, "○ GoSearch: New Query")
	resultsList = append(resultsList, "○ GoSearch: Select Index")
	resultsList = append(resultsList, "○ GoSearch: Crawl and Index")

	prompt := &survey.Select{
		Message: "Results:",
		Options: resultsList,
	}
	elapsed := time.Since(start)

	log.Println("------------------------------------")
	log.Println(util.TerminalCyan+"Queried ", count, " documents in ", elapsed.Milliseconds(), " ms"+util.TerminalReset)
	log.Println("------------------------------------")

	model.ModelLock.Lock()
	status := model.IsComplete
	model.ModelLock.Unlock()
	logStatus(!status, !status, model)

	var selectedLink string

	fmt.Println("------------------------------------------------")
	err = survey.AskOne(prompt, &selectedLink)
	if err != nil {
		log.Fatal(err)
	}

	switch selectedLink {
	case "○ GoSearch: New Query":
		StartQueryPrompt(model)
	case "○ GoSearch: Select Index":
		InitialPrompt(model)
	case "○ GoSearch: Crawl and Index":
		newSite := GetNewWebsitePrompt()
		InitCrawl(newSite, model)
	default:
		cliResponse := formatCliResponse(selectedLink)
		model.ModelLock.Lock()
		fullUrl := model.ReverseUrlFiles[cliResponse]
		// fmt.Println("Full URL:", fullUrl)
		model.ModelLock.Unlock()

		if fullUrl != "" {
			openBrowser(fullUrl)
		}
		startQuery(query, model)
	}

}

// Open the browser to the selected link depending on OS
func openBrowser(link string) {
	fmt.Println(link)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", link)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", link)
	default: // assume Linux or similar
		cmd = exec.Command("xdg-open", link)
	}
	err := cmd.Start()
	if err != nil {
		fmt.Println("Failed to open web browser:", err)
	}
}

// Start the go routine to crawl the website.
func InitCrawl(domain string, model *bm25.Model) {

	_, err := url.ParseRequestURI(domain)
	if err != nil {
		log.Println(util.TerminalRed, "Error parsing URL, please check the domain", util.TerminalReset)
		GetNewWebsitePrompt()
	}

	bm25.ResetModel(model)

	fullUrl, err := url.Parse(domain)
	if err != nil {
		log.Println(util.TerminalRed, "Error parsing URL", util.TerminalReset)
	}

	go func() {
		logStatus(true, true, model)
		webcrawler.CrawlDomainUpdateModel(domain, model, bm25.FileOpsImpl{}, 10000)
		model.ModelLock.Lock()
		model.Name = fullUrl.Host
		model.DA = float32(model.TermCount) / float32(model.DocCount)
		model.IsComplete = true
		model.ModelLock.Unlock()

		logStatus(false, false, model)
	}()

	StartQueryPrompt(model)
}
