package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"
)

const (
	msgFailedToReadFile     = "failed to read file"
	msgFailedToMoveFile     = "failed to move file"
	msgMoveIsStuck          = "move is stuck"
	msgFailedToCreateFolder = "failed to create folder"
	issueFolder             = "_rm_issues"
	pollInterval            = 20 * time.Millisecond
)

// check prints an error message and exits the program if the error is not nil
func check(err error) {
	if err != nil {

		// show the error and stop
		fmt.Println()
		fmt.Println()
		fmt.Println("ERROR:")
		fmt.Fprintln(os.Stderr, err)
		fmt.Println()
		fmt.Println()
		os.Exit(1)
	}
}

// slugify turns a simple error message into a folder name
func slugify(msg string) string {
	var b strings.Builder
	underscore := false
	for _, r := range strings.ToLower(msg) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			underscore = false
			continue
		}
		if !underscore {
			b.WriteByte('_')
			underscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}
