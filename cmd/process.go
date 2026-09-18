package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// processEntries moves scanned files into year/month folders.
func processEntries(source string, addDatePrefix bool) {

	// source folder name prefix used in the folders map
	root := filepath.Base(source)

	// move each scanned file
	for hash, group := range entries {

		// keep the earliest; later copies go to duplicates
		sort.SliceStable(group, func(i, j int) bool {
			return group[i].timestamp.Before(group[j].timestamp)
		})

		earliestPath := ""

		for i, item := range group {

			// build the current file path
			rel, _ := filepath.Rel(root, folders[item.folder])
			origName := item.name
			if item.ext != "" {
				origName += "." + item.ext
			}
			from := filepath.Join(source, rel, origName)
			name := origName
			if addDatePrefix {
				name = withDatePrefix(origName, item.timestamp)
			}

			// duplicate groups: mirror under _mr_issues/duplicates/<hash>
			if len(group) > 1 {
				dup := duplicatePath(source, hash, rel, name)
				if i == 0 {
					_, err := placeCopy(from, dup)
					check(err)
				} else {
					if _, err := placeFile(from, dup); err != nil {
						parkIssue(source, from, name, item, err)
					}
					continue
				}
			}

			// earliest (or only) file goes to the normal destination
			to := destPath(source, item, name)
			path, err := placeFile(from, to)
			if err != nil {
				parkIssue(source, from, name, item, err)
				continue
			}
			if len(group) > 1 && i == 0 {
				earliestPath = path
			}
		}

		if len(group) > 1 && earliestPath != "" {
			writeDuplicateReadme(source, hash, earliestPath)
		}
	}
}

// applyDatePrefixes renames Image and Movie files in place to use a correct date prefix.
// Entry names are updated so later passes see the new paths.
func applyDatePrefixes(source string) {
	root := filepath.Base(source)
	for hash, group := range entries {
		for i, item := range group {
			if (item.kind != Image && item.kind != Movie) || item.issue != "" {
				continue
			}
			rel, _ := filepath.Rel(root, folders[item.folder])
			origName := item.name
			if item.ext != "" {
				origName += "." + item.ext
			}
			newName := withDatePrefix(origName, item.timestamp)
			if newName == origName {
				continue
			}
			from := filepath.Join(source, rel, origName)
			to := filepath.Join(source, rel, newName)
			placed, err := placeFile(from, to)
			check(err)
			base := filepath.Base(placed)
			if item.ext != "" {
				base = strings.TrimSuffix(base, "."+item.ext)
			}
			entries[hash][i].name = base
		}
	}
}

// scannedImages returns paths of Image files found by scan, left in place.
func scannedImages(source string) []string {
	root := filepath.Base(source)
	var paths []string
	for _, group := range entries {
		for _, item := range group {
			if item.kind != Image || item.issue != "" {
				continue
			}
			rel, _ := filepath.Rel(root, folders[item.folder])
			name := item.name
			if item.ext != "" {
				name += "." + item.ext
			}
			paths = append(paths, filepath.Join(source, rel, name))
		}
	}
	return paths
}

// renameViaAI asks whether a filename looks intentional; random names are replaced
// using a vision prompt and renameToSlug. Chat failures are returned to the caller.
// The bool is true when the file was renamed.
func renameViaAI(baseURL, model, apiKey, path string) (bool, error) {

	filename := filepath.Base(path)
	prompt := fmt.Sprintf(loadPrompt("intentional.txt"), filename)
	reply, err := askChat(baseURL, apiKey, model, prompt)
	if err != nil {
		return false, err
	}

	reply = strings.ToLower(strings.TrimSpace(reply))
	if strings.HasPrefix(reply, "yes") {
		return false, nil
	}

	// ask the model for a short descriptive stem based on the image
	suggested, err := askChatImage(baseURL, apiKey, model, loadPrompt("rename.txt"), path)
	if err != nil {
		return false, err
	}
	return renameToSlug(path, suggested)
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

// duplicatePath is the mirrored path under _mr_issues/duplicates/<hash>
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

	_, placeErr := placeFile(from, to)
	check(placeErr)
}

// parkAIFailure moves a file into _mr_issues/ai-failures by media year/month
func parkAIFailure(source, path string) {
	name := filepath.Base(path)
	ts := mediaTimestamp(path)
	to := filepath.Join(source, issueFolder, "ai-failures", ts.Format("2006"), ts.Format("01"), name)
	_, err := placeFile(path, to)
	check(err)
}

// mediaTimestamp returns capture time when available, otherwise mtime (UTC epoch if unreadable)
func mediaTimestamp(path string) time.Time {
	ext := filepath.Ext(path)
	if ext != "" {
		ext = ext[1:]
	}
	ts := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	info, err := os.Stat(path)
	if err == nil {
		ts = info.ModTime()
	}
	if t, ok := imageCaptureTime(path, ext); ok {
		ts = t
	}
	return ts
}
