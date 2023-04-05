package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"os"

	"github.com/deanrtaylor1/gosearch/src/bm25"
	"github.com/deanrtaylor1/gosearch/src/server"
	"github.com/deanrtaylor1/gosearch/src/util"
)

func help() {
	fmt.Println("GoSearch - A simple search engine written in Go")
	fmt.Println("Author: Dean Taylor")
	fmt.Println("Version: 0.1")
	fmt.Println("License: MIT")
	fmt.Println("default start: gosearch.exe launches search engine and crawler on localhost:8080")

	fmt.Println("CLI Usage: PROGRAM [SUBCOMMAND] [OPTIONS]")
	fmt.Println("----------------------------------")
	fmt.Println("Subcommands:")
	fmt.Println("    cli:                            start server with cli interface")
	fmt.Println("    help:                            list all commands")

}

func openBrowser() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "http://localhost:8080")
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "http://localhost:8080")
	default: // assume Linux or similar
		cmd = exec.Command("xdg-open", "http://localhost:8080")
	}
	err := cmd.Start()
	if err != nil {
		fmt.Println("Failed to open web browser:", err)
	}
}

func main() {
	if len(os.Args) < 1 {
		help()
		os.Exit(1)
	}
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println(util.TerminalCyan + "Initializing server with empty model" + util.TerminalReset)
		model := bm25.NewEmptyModel()
		openBrowser()
		server.Serve(model)
	}
	program := args[0]

	switch program {
	case "cli":

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

	case "-help":
		help()

	default:
		help()
	}

}
