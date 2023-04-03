package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
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
	case "start":
		selectedDirectory := strings.Replace(util.SelectDirectory(), "○ ", "", -1)
		fmt.Printf("Selected directory: %s\n", selectedDirectory)
		fmt.Println("Indexing with bm25")

		model := &bm25.Model{
			Name:      selectedDirectory,
			TFPD:      make(bm25.TermFreqPerDoc),
			DF:        make(bm25.DocFreq),
			ModelLock: &sync.Mutex{},
		}
		go func() {
			bm25.AddFolderToModel("./"+selectedDirectory, model)
			model.ModelLock.Lock()
			model.DA = float32(model.TermCount) / float32(model.DocCount)
			fmt.Println(model.DA)
			model.ModelLock.Unlock()
		}()

		server.Serve(model)
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
				fmt.Println("Indexing with bm25")
				model := &bm25.Model{
					TFPD: make(bm25.TermFreqPerDoc),
					DF:   make(bm25.DocFreq),
				}
				bm25.AddFolderToModel(dirPath, model)
				model.DA = float32(model.TermCount) / float32(model.DocCount)
				bm25.ModelToJSON(*model, true, "index.json")

			default:
				fmt.Println("Invalid algorithm")
				help()
			}
		} else {
			fmt.Println("Indexing with bm25")
			model := &bm25.Model{
				TFPD: make(bm25.TermFreqPerDoc),
				DF:   make(bm25.DocFreq),
			}
			bm25.AddFolderToModel(dirPath, model)
			model.DA = float32(model.TermCount) / float32(model.DocCount)
			bm25.ModelToJSON(*model, true, "index.json")
		}

	case "search":
		if len(args) != 2 {
			log.Fatal("Path to folder must be provided.")
		}
		indexPath := args[1]
		tfidf.CheckIndex(indexPath)

	case "help":
		help()
	case "server":
		if len(args) < 2 {
			log.Fatal("Path to folder must be provided.")
		}
		indexPath := args[1]
		if len(args) > 2 && args[2] == "-a" {

			switch args[3] {
			// case "tfidf":
			// 	f, err := os.Open(indexPath)
			// 	if err != nil {
			// 		log.Fatal(err)

			// 	}
			// 	defer f.Close()

			// 	var model tfidf.Model

			// 	err = json.NewDecoder(f).Decode(&model)
			// 	if err != nil {
			// 		log.Fatal(err)

			// 	}

			// 	server.Serve(model)
			case "bm25":
				f, err := os.Open(indexPath)
				if err != nil {
					log.Fatal(err)

				}
				defer f.Close()

				var model bm25.Model

				err = json.NewDecoder(f).Decode(&model)
				if err != nil {
					log.Fatal(err)

				}

				server.Serve(&model)
			default:
				fmt.Println("Invalid algorithm")
				help()

			}
		} else {
			log.Fatal("Algorithm must be provided.")
		}

	case "crawl":
		if len(args) != 2 {
			help()
			log.Fatal("Domain must be provided.")

		}
		domain := args[1]
		webcrawler.CrawlDomain(domain)
	default:
		help()
	}

}
