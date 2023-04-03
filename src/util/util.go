package util

import (
	"encoding/json"
	"fmt"

	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
)

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
	files, err := os.ReadDir(".")
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

func GetDirLength(dirName string) int {
	files, err := os.ReadDir("./" + dirName)
	if err != nil {
		log.Fatal(err)
	}

	return len(files)
}
