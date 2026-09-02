package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// sourceFolder returns the path to the source folder
func sourceFolder() string {
	if len(os.Args) != 2 {
		check(fmt.Errorf("expected <folder>"))
	}

	path, err := filepath.Abs(os.Args[1])
	check(err)

	info, err := os.Stat(path)
	check(err)
	if !info.IsDir() {
		check(fmt.Errorf("not a folder: %s", path))
	}

	return path
}
