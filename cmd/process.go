package main

import (
	"path/filepath"
	"sort"
)

// processEntries moves scanned files into year/month folders
func processEntries(source string) {

	// source folder name prefix used in the folders map
	root := filepath.Base(source)

	// move each scanned file
	for _, group := range entries {

		// keep the earliest; later copies go to duplicates
		sort.SliceStable(group, func(i, j int) bool {
			return group[i].timestamp.Before(group[j].timestamp)
		})

		for i, item := range group {

			// build the current file path
			rel, _ := filepath.Rel(root, folders[item.folder])
			name := item.name
			if item.ext != "" {
				name += "." + item.ext
			}
			from := filepath.Join(source, rel, name)

			// build the destination path
			to := destPath(source, item, name, i > 0)
			if err := placeFile(from, to); err != nil {
				parkIssue(source, from, name, item, err)
			}
		}
	}
}

// destPath is the year/month path for a scanned file
func destPath(source string, item entry, name string, duplicate bool) string {
	if item.issue != "" {
		return filepath.Join(source, "issues", "errors", slugify(item.issue), item.year, item.month, name)
	}
	if duplicate {
		return filepath.Join(source, "issues", "duplicates", item.year, item.month, name)
	}
	if item.kind == Other {
		return filepath.Join(source, "issues", "other_filetypes", item.year, item.month, name)
	}
	return filepath.Join(source, item.year, item.month, name)
}

// parkIssue moves a failed file into issues/errors
func parkIssue(source, from, name string, item entry, err error) {

	// already heading for an errors folder, so stop
	to := filepath.Join(source, "issues", "errors", slugify(err.Error()), item.year, item.month, name)
	if to == destPath(source, item, name, false) {
		check(err)
	}

	check(placeFile(from, to))
}
