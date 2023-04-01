package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"os"

	"github.com/deanrtaylor1/gosearch/src/bm25"
	"github.com/deanrtaylor1/gosearch/src/server"
	"github.com/deanrtaylor1/gosearch/src/tfidf"
	"github.com/deanrtaylor1/gosearch/src/util"

	webcrawler "github.com/deanrtaylor1/gosearch/src/web-crawler"
)

func help() {
	fmt.Println("Usage: PROGRAM [SUBCOMMAND] [OPTIONS]")
	fmt.Println("----------------------------------")
	fmt.Println("    index:                           index -a <algorithm> <path to folder>")
	fmt.Println("                                     Available algorithms: tfidf, bm25")
	fmt.Println("    search:                          search <path to folder> <query>")
	fmt.Println("    help:                            list all commands")
	fmt.Println("    serve:                           start local http server")
}

func main() {

	if len(os.Args) < 1 {
		help()
		os.Exit(1)
	}

	args := os.Args[1:]

	if len(args) < 1 {
		help()
		os.Exit(1)
	}
	program := args[0]

	switch program {
	case "index":

		if len(args) < 2 {
			help()
			os.Exit(1)
		}
		dirPath := args[1]

		if len(args) > 2 && args[2] == "-a" {

			switch args[3] {
			case "tfidf":
				fmt.Println("Indexing with tfidf")
				model := &tfidf.Model{
					TFPD: make(tfidf.TermFreqPerDoc),
					DF:   make(tfidf.DocFreq),
				}
				tfidf.AddFolderToModel(dirPath, model)
				tfidf.ModelToJSON(*model, true, "index.json")

			case "bm25":
				fmt.Println("TODO")
				fmt.Println("Indexing with bm25")
				model := &bm25.Model{
					TFPD: make(bm25.TermFreqPerDoc),
					DF:   make(bm25.DocFreq),
				}
				bm25.AddFolderToModel(dirPath, model)
				bm25.ModelToJSON(*model, true, "index.json")

			default:
				fmt.Println("Invalid algorithm")
				help()

			}
		} else {
			fmt.Println("No flag found, indexing with tfidf")
			model := &tfidf.Model{
				TFPD: make(tfidf.TermFreqPerDoc),
				DF:   make(tfidf.DocFreq),
			}
			tfidf.AddFolderToModel(dirPath, model)
			tfidf.ModelToJSON(*model, true, "index.json")
		}

	case "search":
		if len(args) != 2 {
			log.Fatal("Path to folder must be provided.")
		}
		indexPath := args[1]
		tfidf.CheckIndex(indexPath)
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

		var model tfidf.Model

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
		util.MapToJSON(urls, true, dirName+"/urls.json")
	default:
		help()
	}

}
