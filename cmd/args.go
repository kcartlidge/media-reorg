package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// sourceFolder returns the path to the source folder
func sourceFolder() string {

	// require a single folder argument
	if len(os.Args) != 2 {
		check(fmt.Errorf("expected <folder>"))
	}

	// resolve to an absolute path
	path, err := filepath.Abs(os.Args[1])
	check(err)

	// ensure it exists and is a folder
	info, err := os.Stat(path)
	check(err)
	if !info.IsDir() {
		check(fmt.Errorf("not a folder: %s", path))
	}

	return path
}
