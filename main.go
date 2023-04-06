package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"os"

	"github.com/deanrtaylor1/gosearch/bm25"
	"github.com/deanrtaylor1/gosearch/server"
	"github.com/deanrtaylor1/gosearch/util"
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

func downloadStaticDir() error {
	staticDir := "static"

	// Check if the static directory exists
	_, err := os.Stat(staticDir)
	if !os.IsNotExist(err) {
		// Directory exists, no need to download
		return nil
	}

	// Create the static directory
	err = os.Mkdir(staticDir, 0755)
	if err != nil {
		return err
	}

	// List of files in the static directory on GitHub
	files := []string{
		"index.html",
		"styles.css",
		"favicon.ico",
		"index.js",
	}

	// Base URL for the raw content on GitHub
	baseURL := "https://raw.githubusercontent.com/DeanRTaylor1/gosearch/main/static/"

	// Download each file
	for _, file := range files {
		resp, err := http.Get(baseURL + file)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Create the file on disk
		out, err := os.Create(filepath.Join(staticDir, file))
		if err != nil {
			return err
		}
		defer out.Close()

		// Copy the content from the response to the file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	files, err := os.ReadDir("./static")
	if err != nil || len(files) != 4 {
		log.Println(err)
		err := downloadStaticDir()
		if err != nil {
			fmt.Printf("Error downloading the static directory: %v\n", err)
			fmt.Println("Please make sure you have write permissions in the current directory.")
		}
	}
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
				bm25.LoadCachedGobToModel("./indexes/"+selectedDirectory, model)
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
