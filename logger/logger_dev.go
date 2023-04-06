//go:build dev
// +build dev

package logger

import "fmt"

func HandleError(err error) {
	fmt.Printf("Dev Mode - Error: %v\n", err)
}
