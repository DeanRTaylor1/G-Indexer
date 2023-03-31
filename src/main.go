package main

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"os"

	"github.com/deanrtaylor1/gosearch/src/lexer"
	"github.com/deanrtaylor1/gosearch/src/server"
	"github.com/deanrtaylor1/gosearch/src/types"
	webcrawler "github.com/deanrtaylor1/gosearch/src/web-crawler"
)

func addFolderToModel(dirPath string, model *types.Model) {

	dir, err := os.Open(dirPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}

	for _, fi := range fileInfos {
		if fi.IsDir() {
			if fi.Name() == "specs" {
				continue
			}
			subDirPath := filepath.Join(dirPath, fi.Name())
			addFolderToModel(subDirPath, model)
		}
		switch filepath.Ext(fi.Name()) {
		case ".xhtml", ".xml":
			filePath := dirPath + "/" + fi.Name()
			fmt.Println("Indexing file: ", filePath)
			content := lexer.ReadEntireXMLFile(filePath)
			fileSize := len(content)

			fmt.Println(filePath, " => ", fileSize)
			tf := make(types.TermFreq)

			lexer := lexer.NewLexer(content)
			for {
				token, err := lexer.Next()
				if err != nil {
					fmt.Println("EOF")
					break
				}

				tf[token] += 1
				//stats := mapToSortedSlice(tf)
			}
			for token := range tf {
				model.DF[token] += 1
			}

			model.TFPD[filePath] = ConvertToDocData(tf)

		case ".html":
			fmt.Println("TODO IMPLEMENT HTML PARSER")
			filePath := dirPath + "/" + fi.Name()
			fmt.Println("Indexing file: ", filePath)
			content := lexer.ReadEntireHTMLFile(filePath)
			fileSize := len(content)

			fmt.Println(filePath, " => ", fileSize)
			tf := make(types.TermFreq)

			lexer := lexer.NewLexer(content)
			for {
				token, err := lexer.Next()
				if err != nil {
					fmt.Println("EOF")
					break
				}

				tf[token] += 1
				//stats := mapToSortedSlice(tf)
			}
			for token := range tf {
				model.DF[token] += 1
			}

			model.TFPD[filePath] = ConvertToDocData(tf)

		default:
			fmt.Fprint(os.Stderr, "\033[31mSkipping file:", fi.Name(), "(not HTML. .xhtml or .xml)\033[0m")
			fmt.Println()
			continue
		}
	}

}

func ConvertToDocData(tf types.TermFreq) types.DocData {
	var termCount int

	for _, freq := range tf {
		termCount += freq

	}

	docData := &types.DocData{
		TermCount: termCount,
		Terms:     tf,
	}
	return *docData
}

func help() {
	fmt.Println("Usage: PROGRAM [SUBCOMMAND] [OPTIONS]")
	fmt.Println("----------------------------------")
	fmt.Println("    index:                           index <path to folder>")
	fmt.Println("    search:                          search <path to folder> <query>")
	fmt.Println("    help:                            list all commands")
	fmt.Println("    serve:                           start local http server")

}

func main() {

	args := os.Args[1:]
	if len(args) < 1 {
		help()
		os.Exit(1)
	}
	program := args[0]

	switch program {
	case "index":
		if len(args) != 2 {
			log.Fatal("Path to folder must be provided.")
		}
		dirPath := args[1]
		model := &types.Model{
			TFPD: make(types.TermFreqPerDoc),
			DF:   make(types.DocFreq),
		}
		addFolderToModel(dirPath, model)
		lexer.ModelToJSON(*model, true, "index.json")
		// termFreqIndex := tfIndexFolder(dirPath)
		// Lexer.MapToJSON(termFreqIndex, true, "index.json")
	case "search":
		if len(args) != 2 {
			log.Fatal("Path to folder must be provided.")
		}
		indexPath := args[1]
		lexer.CheckIndex(indexPath)
		fmt.Println("TODO: IMPLEMENT SEARCH FUNCTION")
	case "help":
		help()
	case "server":
		if len(args) != 2 {
			log.Fatal("Path to folder must be provided.")
		}
		indexPath := args[1]
		//tfIndex, err := Lexer.CheckIndex(indexPath)
		f, err := os.Open(indexPath)
		if err != nil {
			log.Fatal(err)

		}
		defer f.Close()

		var model types.Model

		err = json.NewDecoder(f).Decode(&model)
		if err != nil {
			log.Fatal(err)

		}

		server.Serve(model)
	case "crawl":
		if len(args) != 2 {
			help()
			log.Fatal("Path to folder must be provided.")

		}
		domain := args[1]
		fmt.Println("crawling domain: ", domain)
		visited := make(map[string]bool)
		urls := make(map[string]string)

		visitedMutex := sync.Mutex{}
		dirName := webcrawler.Crawl(domain, domain, nil, true, &visitedMutex, &visited, &urls)

		visitedMutex.Lock()
		defer visitedMutex.Unlock()
		lexer.MapToJSON(urls, true, dirName+"/urls.json")
	default:
		help()
	}

}
