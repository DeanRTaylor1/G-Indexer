package main

import (
	"fmt"
	"log"
	"path/filepath"

	"os"

	"github.com/deanrtaylor1/gosearch/src/Lexer"
	"github.com/deanrtaylor1/gosearch/src/Types"
	"github.com/deanrtaylor1/gosearch/src/server"
)

func tfIndexFolder(dirPath string) Types.TermFreqIndex {
	dir, err := os.Open(dirPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}

	termFreqIndex := make(Types.TermFreqIndex)

	for _, fi := range fileInfos {
		if fi.IsDir() {
			if fi.Name() == "specs" {
				continue
			}
			subDirPath := filepath.Join(dirPath, fi.Name())
			subTermFreqIndex := tfIndexFolder(subDirPath)
			for k, v := range subTermFreqIndex {
				termFreqIndex[k] = v
			}
			continue
		}
		if filepath.Ext(fi.Name()) != ".xhtml" && filepath.Ext(fi.Name()) != ".xml" {
			continue
		}
		filePath := dirPath + "/" + fi.Name()
		fmt.Println("Indexing file: ", filePath)
		content := Lexer.ReadEntireXMLFile(filePath)
		fileSize := len(content)

		fmt.Println(filePath, " => ", fileSize)
		tf := make(Types.TermFreq)

		lexer := Lexer.NewLexer(content)
		for {
			token, err := lexer.Next()
			if err != nil {
				fmt.Println("EOF")
				break
			}
			if _, ok := tf[token]; ok {
				tf[token] += 1
			} else {
				tf[token] = 1
			}
			//stats := mapToSortedSlice(tf)
		}

		termFreqIndex[filePath] = tf
	}
	return termFreqIndex

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
		termFreqIndex := tfIndexFolder(dirPath)
		Lexer.MapToJSON(termFreqIndex, true, "index.json")
	case "search":
		if len(args) != 2 {
			log.Fatal("Path to folder must be provided.")
		}
		indexPath := args[1]
		Lexer.CheckIndex(indexPath)
		fmt.Println("TODO: IMPLEMENT SEARCH FUNCTION")
	case "help":
		help()
	case "server":
		if len(args) != 2 {
			log.Fatal("Path to folder must be provided.")
		}
		indexPath := args[1]
		tfIndex, err := Lexer.CheckIndex(indexPath)

		if err != nil {
			log.Fatal(err)
		}

		// for k, v := range tfIndex {
		// 	fmt.Println(k)
		// 	for k2, v2 := range v {
		// 		fmt.Println(k2, v2)
		// 	}
		// }

		server.Serve(tfIndex)
		fmt.Println("TODO: IMPLEMENT SERVER FUNCTION")
	default:
		help()
	}
	/* for p, tf := range termFreqIndex {*/
	/*fmt.Printf("%v has %v unique terms", p, len(tf))*/
	/*fmt.Println()*/
	/*}*/

}
