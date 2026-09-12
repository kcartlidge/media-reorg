package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// kind is Image, Movie, or Other
type kind int

// Image, Movie, and Other classify a file by extension
const (
	Image kind = iota
	Movie
	Other
)

// entry is a scanned file
type entry struct {
	folder    int
	name      string
	ext       string
	kind      kind
	timestamp time.Time
	size      int64
	year      string
	month     string
	issue     string
}

// folders maps auto-incrementing ids to folder paths
var folders map[int]string

// entries maps file hashes to one or more scanned files
var entries map[string][]entry

// scan walks the source folder and populates the folders and entries maps
func scan(source string) {

	// prepare folder and file maps
	folders = make(map[int]string)
	entries = make(map[string][]entry)
	ids := make(map[string]int)
	next := 1
	root := filepath.Base(source)
	issueSkipped := 0

	// walk every folder and file
	err := filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && !d.IsDir() && !skipJunk(d.Name()) {
				addFile(path, d, ids)
			}
			return nil
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return nil
		}

		// skip the issue folder from a previous run
		if underIssueFolder(rel) {
			if d.IsDir() && rel == issueFolder {
				issueSkipped = countFilesInTree(path)
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {

			// record the folder
			name := filepath.Join(root, rel)
			folders[next] = name
			ids[path] = next
			next++

			// print folders one or two levels deep
			if rel != "." {
				depth := strings.Count(rel, string(filepath.Separator)) + 1
				if depth <= 2 {
					if depth == 2 && hasSubfolder(path) {
						name += " (+)"
					}
					fmt.Println(name)
				}
			}
			return nil
		}

		// ignore OS metadata files
		if skipJunk(d.Name()) {
			return nil
		}

		addFile(path, d, ids)
		return nil
	})
	check(err)

	// report skipped issue files
	if issueSkipped > 0 {
		fmt.Println("Skipped", issueSkipped, "files in", issueFolder+".")
	}
}

// underIssueFolder reports whether rel is inside the issue folder tree
func underIssueFolder(rel string) bool {
	return rel == issueFolder || strings.HasPrefix(rel, issueFolder+string(filepath.Separator))
}

// countFilesInTree counts non-junk files under root
func countFilesInTree(root string) int {
	n := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && !skipJunk(d.Name()) {
			n++
		}
		return nil
	})
	check(err)
	return n
}

// addFile records a scanned file, marking read failures as issues
func addFile(path string, d fs.DirEntry, ids map[string]int) {

	// split the filename
	filename := d.Name()
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	if ext != "" {
		ext = ext[1:]
	}

	// read timestamps and size
	issue := ""
	ts := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	var size int64
	info, err := d.Info()
	if err != nil {
		issue = msgFailedToReadFile
	} else {
		ts = info.ModTime()
		size = info.Size()
	}

	// hash the contents
	hash, err := fileHash(path)
	if err != nil {
		issue = msgFailedToReadFile
		hash = path
	}

	entries[hash] = append(entries[hash], entry{
		folder:    ids[filepath.Dir(path)],
		name:      base,
		ext:       ext,
		kind:      kindFromExt(ext),
		timestamp: ts,
		size:      size,
		year:      ts.Format("2006"),
		month:     ts.Format("01"),
		issue:     issue,
	})
}

// fileHash returns the SHA-256 hex digest of the file at path
func fileHash(path string) (string, error) {

	// open the file
	file, err := os.Open(path)
	if err != nil {
		return "", errors.New(msgFailedToReadFile)
	}
	defer file.Close()

	// hash the contents
	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", errors.New(msgFailedToReadFile)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// kindFromExt returns Image, Movie, or Other for a file extension
func kindFromExt(ext string) kind {
	switch strings.ToLower(ext) {
	case "jpg", "jpeg", "png", "gif", "webp", "heic", "heif", "tif", "tiff", "bmp", "raw", "dng", "cr2", "nef", "arw", "avif":
		return Image
	case "mp4", "mov", "avi", "mkv", "m4v", "wmv", "mpg", "mpeg", "webm", "3gp", "mts", "m2ts":
		return Movie
	default:
		return Other
	}
}

// hasSubfolder reports whether path contains any immediate subfolders
func hasSubfolder(path string) bool {

	// read the folder contents
	items, err := os.ReadDir(path)
	check(err)

	// look for an immediate subfolder
	for _, item := range items {
		if item.IsDir() {
			return true
		}
	}
	return false
}

// skipJunk reports whether name is OS metadata that should not be organised
func skipJunk(name string) bool {
	switch strings.ToLower(name) {
	case ".ds_store", "thumbs.db", "ehthumbs.db", "desktop.ini", ".directory":
		return true
	}
	return strings.HasPrefix(name, "._")
}

// placeFile creates the destination folder and moves the file
func placeFile(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
		return errors.New(msgFailedToCreateFolder)
	}
	return moveFile(from, to)
}

// moveFile starts a rename and waits until it appears, errors, or times out
func moveFile(from, to string) error {
	if from == to {
		return nil
	}

	// nothing to move if the source has already gone
	if _, err := os.Stat(from); os.IsNotExist(err) {
		if _, destErr := os.Stat(to); destErr == nil {
			return nil
		}
		return errors.New(msgFailedToMoveFile)
	}

	// start the move
	named := make(chan string, 1)
	done := make(chan error, 1)
	go func() {
		dest := freePath(to)
		named <- dest
		done <- os.Rename(from, dest)
	}()
	to = <-named

	// wait until this file has moved, an error occurs, or a timeout
	deadline := time.Now().Add(20 * time.Second)
	for {
		select {
		case err := <-done:
			if err != nil {
				return errors.New(msgFailedToMoveFile)
			}
		default:
		}

		_, destErr := os.Stat(to)
		_, srcErr := os.Stat(from)
		if destErr == nil && os.IsNotExist(srcErr) {
			return nil
		}

		if time.Now().After(deadline) {
			return errors.New(msgMoveIsStuck)
		}

		time.Sleep(pollInterval)
	}
}

// freePath returns to, or to with _N before the extension if to already exists
func freePath(to string) string {

	// keep the original name when it is free
	if _, err := os.Stat(to); os.IsNotExist(err) {
		return to
	}

	// find the highest existing _N variant in the same folder
	dir := filepath.Dir(to)
	base := filepath.Base(to)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	next := 1
	prefix := name + "_"
	items, err := os.ReadDir(dir)
	if err == nil {
		for _, item := range items {
			fn := item.Name()
			if !strings.HasPrefix(fn, prefix) {
				continue
			}
			if ext != "" && !strings.HasSuffix(fn, ext) {
				continue
			}
			mid := strings.TrimSuffix(strings.TrimPrefix(fn, prefix), ext)
			n, convErr := strconv.Atoi(mid)
			if convErr != nil || n < next {
				continue
			}
			next = n + 1
		}
	}

	// use the next free numbered name
	for {
		candidate := filepath.Join(dir, name+"_"+strconv.Itoa(next)+ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
		next++
	}
}

// clearup removes empty folders and folders that only contain junk files
func clearup(source string) {
	walkClear(source, source)
}

// walkClear visits folders depth-first and removes empty or junk-only ones
func walkClear(path, source string) {

	// clear child folders first
	items, err := os.ReadDir(path)
	check(err)
	for _, item := range items {
		if item.IsDir() {
			walkClear(filepath.Join(path, item.Name()), source)
		}
	}

	// never remove the source folder itself
	if path == source {
		return
	}

	// remove this folder if it is empty or only junk
	items, err = os.ReadDir(path)
	check(err)
	if !emptyOrJunkOnly(items) {
		return
	}
	removePath(path)
}

// emptyOrJunkOnly reports whether items has no real files or folders
func emptyOrJunkOnly(items []os.DirEntry) bool {
	for _, item := range items {
		if item.IsDir() || !skipJunk(item.Name()) {
			return false
		}
	}
	return true
}

// removePath starts a delete and waits until it is gone, errors, or times out
func removePath(path string) {

	// start the removal
	done := make(chan error, 1)
	go func() {
		done <- os.RemoveAll(path)
	}()

	// wait until the path is gone, an error occurs, or a timeout
	deadline := time.Now().Add(20 * time.Second)
	for {
		select {
		case err := <-done:
			check(err)
		default:
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			return
		}

		if time.Now().After(deadline) {
			check(fmt.Errorf("%s is stuck", path))
		}

		time.Sleep(pollInterval)
	}
}
