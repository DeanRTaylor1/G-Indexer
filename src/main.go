package main

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"os"

	"github.com/deanrtaylor1/gosearch/src/Lexer"
	"github.com/deanrtaylor1/gosearch/src/Types"
	"github.com/deanrtaylor1/gosearch/src/server"
)

func addFolderToModel(dirPath string, model *Types.Model) {

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
		if filepath.Ext(fi.Name()) != ".xhtml" && filepath.Ext(fi.Name()) != ".xml" {
			fmt.Fprint(os.Stderr, "\033[31mSkipping file:", fi.Name(), "(not .xhtml or .xml)\033[0m")
			fmt.Println()
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

			tf[token] += 1
			//stats := mapToSortedSlice(tf)
		}
		for token := range tf {
			model.DF[token] += 1
		}

		model.TFPD[filePath] = ConvertToDocData(tf)
	}

}

func ConvertToDocData(tf Types.TermFreq) Types.DocData {
	var termCount int

	for _, freq := range tf {
		termCount += freq

	}

	docData := &Types.DocData{
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
		model := &Types.Model{
			TFPD: make(Types.TermFreqPerDoc),
			DF:   make(Types.DocFreq),
		}
		addFolderToModel(dirPath, model)
		Lexer.ModelToJSON(*model, true, "index.json")
		// termFreqIndex := tfIndexFolder(dirPath)
		// Lexer.MapToJSON(termFreqIndex, true, "index.json")
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
		//tfIndex, err := Lexer.CheckIndex(indexPath)
		f, err := os.Open(indexPath)
		if err != nil {
			log.Fatal(err)

		}
		defer f.Close()

		var model Types.Model

		err = json.NewDecoder(f).Decode(&model)
		if err != nil {
			log.Fatal(err)

		}

		// for k, v := range tfIndex {
		// 	fmt.Println(k)
		// 	for k2, v2 := range v {
		// 		fmt.Println(k2, v2)
		// 	}
		// }

		server.Serve(model)
		fmt.Println("TODO: IMPLEMENT SERVER FUNCTION")
	default:
		help()
	}
	/* for p, tf := range termFreqIndex {*/
	/*fmt.Printf("%v has %v unique terms", p, len(tf))*/
	/*fmt.Println()*/
	/*}*/

}
