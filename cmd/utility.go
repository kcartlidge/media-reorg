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
	msgFailedToCopyFile     = "failed to copy file"
	msgCopyIsStuck          = "copy is stuck"
	msgFailedToCreateFolder = "failed to create folder"
	issueFolder             = "_rm_issues"
	pollInterval            = 20 * time.Millisecond
	pollTimeout             = 20 * time.Second
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
	return slugifyWith(msg, '_')
}

// slugifyFilename turns suggested text into a hyphenated filename stem
func slugifyFilename(msg string) string {
	return slugifyWith(msg, '-')
}

// slugifyWith turns text into a slug using sep between word pieces
func slugifyWith(msg string, sep byte) string {
	var b strings.Builder
	pending := false
	for _, r := range strings.ToLower(msg) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			pending = false
			continue
		}
		if !pending {
			b.WriteByte(sep)
			pending = true
		}
	}
	return strings.Trim(b.String(), string(sep))
}
