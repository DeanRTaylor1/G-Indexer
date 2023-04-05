package main

import (
	"fmt"
	"strings"

	"os"

	"github.com/deanrtaylor1/gosearch/src/bm25"
	"github.com/deanrtaylor1/gosearch/src/server"
	"github.com/deanrtaylor1/gosearch/src/util"
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

		model := bm25.NewEmptyModel()
		if selectedDirectory == "Start server" {
			fmt.Println("Starting server with no Index")

		} else {
			fmt.Println("Starting server and indexing directory: ", selectedDirectory)
			model.Name = selectedDirectory
			go func() {
				bm25.LoadCachedGobToModel("./"+selectedDirectory, model)
				model.ModelLock.Lock()
				model.DA = float32(model.TermCount) / float32(model.DocCount)
				model.IsComplete = true
				model.ModelLock.Unlock()
			}()

		}
		server.Serve(model)

	case "help":
		help()

	default:
		help()
	}

}
