package util

import (
	"encoding/json"
	"fmt"
	"os"
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
