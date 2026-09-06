package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// processEntries moves scanned files into year/month folders
func processEntries(source string) {

	// source folder name prefix used in the folders map
	root := filepath.Base(source)

	// move each scanned file
	for _, group := range entries {
		for _, item := range group {

			// build the current file path
			rel, _ := filepath.Rel(root, folders[item.folder])
			name := item.name
			if item.ext != "" {
				name += "." + item.ext
			}
			from := filepath.Join(source, rel, name)

			// build the destination path
			destDir := filepath.Join(source, item.year, item.month)
			if item.kind == Other {
				destDir = filepath.Join(source, "issues", "others", item.year, item.month)
			}
			to := filepath.Join(destDir, name)

			// move the file
			check(os.MkdirAll(destDir, 0755))
			moveFile(from, to)
		}
	}
}

// moveFile starts a rename and waits until it appears, errors, or times out
func moveFile(from, to string) {
	if from == to {
		return
	}

	// start the move
	done := make(chan error, 1)
	go func() {
		done <- os.Rename(from, to)
	}()

	// wait until this file has moved, an error occurs, or a timeout
	deadline := time.Now().Add(20 * time.Second)
	for {
		select {
		case err := <-done:
			check(err)
		default:
		}

		_, destErr := os.Stat(to)
		_, srcErr := os.Stat(from)
		if destErr == nil && os.IsNotExist(srcErr) {
			return
		}

		if time.Now().After(deadline) {
			check(fmt.Errorf("%s is stuck", from))
		}

		time.Sleep(200 * time.Millisecond)
	}
}
