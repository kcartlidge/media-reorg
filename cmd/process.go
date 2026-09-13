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
	for hash, group := range entries {

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

			// duplicate groups: mirror under _rm_issues/duplicates/<hash>
			if len(group) > 1 {
				dup := duplicatePath(source, hash, rel, name)
				if i == 0 {
					check(placeCopy(from, dup))
				} else {
					if err := placeFile(from, dup); err != nil {
						parkIssue(source, from, name, item, err)
					}
					continue
				}
			}

			// earliest (or only) file goes to the normal destination
			to := destPath(source, item, name)
			if err := placeFile(from, to); err != nil {
				parkIssue(source, from, name, item, err)
			}
		}
	}
}

// destPath is the year/month path for a scanned file
func destPath(source string, item entry, name string) string {
	if item.issue != "" {
		return filepath.Join(source, issueFolder, "errors", slugify(item.issue), item.year, item.month, name)
	}
	if item.kind == Other {
		return filepath.Join(source, issueFolder, "other_filetypes", item.year, item.month, name)
	}
	return filepath.Join(source, item.year, item.month, name)
}

// duplicatePath is the mirrored path under _rm_issues/duplicates/<hash>
func duplicatePath(source, hash, rel, name string) string {
	return filepath.Join(source, issueFolder, "duplicates", hash, rel, name)
}

// parkIssue moves a failed file into the issues/errors tree
func parkIssue(source, from, name string, item entry, err error) {

	// already heading for an errors folder, so stop
	to := filepath.Join(source, issueFolder, "errors", slugify(err.Error()), item.year, item.month, name)
	if to == destPath(source, item, name) {
		check(err)
	}

	check(placeFile(from, to))
}
