package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// scan walks the source folder and prints folders 1 or 2 levels deep
func scan(source string) {
	root := filepath.Base(source)
	err := filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		check(err)
		if !d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(source, path)
		check(err)
		if rel == "." {
			return nil
		}
		depth := strings.Count(rel, string(filepath.Separator)) + 1
		if depth > 2 {
			return nil
		}
		name := filepath.Join(root, rel)
		if depth == 2 && hasSubfolder(path) {
			name += " (+)"
		}
		fmt.Println(name)
		return nil
	})
	check(err)
}

func hasSubfolder(path string) bool {
	entries, err := os.ReadDir(path)
	check(err)
	for _, entry := range entries {
		if entry.IsDir() {
			return true
		}
	}
	return false
}
