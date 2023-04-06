package util

import (
	"encoding/json"
	"fmt"

	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
)

type IndexedData struct {
	URL     string
	Content string // Or any other data structure used for storing indexed content
}

func JSONToFile(j []byte, filename string) {
	fmt.Println("j length:", len(j)) // debugging line
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	l, err := f.Write(j)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func MapToJSON(m map[string]string, createFile bool, filename string) string {
	// Convert map[string]string to map[string]interface{}
	mi := make(map[string]interface{}, len(m))
	for k, v := range m {
		mi[k] = v
	}

	return MapToJSONGeneric(mi, createFile, filename)
}

func MapToJSONGeneric(m map[string]interface{}, createFile bool, filename string) string {
	if len(m) == 0 {
		fmt.Println("map is empty")
		return ""
	}

	b, err := json.Marshal(m)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}
	if createFile {
		JSONToFile(b, filename)
	}
	return string(b)
}

func SelectDirectory() string {
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
	directories = append(directories, "○ Start server")

	prompt := &survey.Select{
		Message: "Select a directory to index:",
		Options: directories,
	}

	var selectedDirectory string
	err = survey.AskOne(prompt, &selectedDirectory)
	if err != nil {
		log.Fatal(err)
	}

	return selectedDirectory
}

func GetCurrentAvailableModelDirectories() []string {
	files, err := os.ReadDir("./indexes")
	if err != nil {
		log.Println(err)
		err := os.Mkdir("./indexes", os.FileMode(0777))
		if err != nil {
			log.Println(err)
		}
	}

	directories := []string{}
	for _, f := range files {
		if f.IsDir() {
			if f.Name() == "src" || f.Name() == ".git" {
				continue
			}
			directories = append(directories, f.Name())
		}
	}

	return directories
}

func CheckDirIsValid(dirName string) (bool, error) {
	_, err := os.Stat("./" + dirName)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Directory does not exist
		}
		return false, err // Some other error occurred
	}
	return true, nil // Directory exists
}

func GetDirLength(dirName string) int {
	files, err := os.ReadDir("./" + dirName)
	if err != nil {
		log.Fatal(err)
	}

	return len(files)
}

const (
	TerminalReset  = "\033[0m"
	TerminalRed    = "\033[31m"
	TerminalGreen  = "\033[32m"
	TerminalYellow = "\033[33m"
	TerminalBlue   = "\033[34m"
	TerminalPurple = "\033[35m"
	TerminalCyan   = "\033[36m"
	TerminalWhite  = "\033[37m"
)
