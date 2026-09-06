package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

	// walk every folder and file
	err := filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		check(err)
		rel, err := filepath.Rel(source, path)
		check(err)

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

		// split the filename
		info, err := d.Info()
		check(err)
		filename := d.Name()
		ext := filepath.Ext(filename)
		base := strings.TrimSuffix(filename, ext)
		if ext != "" {
			ext = ext[1:]
		}

		// record the file
		ts := info.ModTime()
		hash := fileHash(path)
		entries[hash] = append(entries[hash], entry{
			folder:    ids[filepath.Dir(path)],
			name:      base,
			ext:       ext,
			kind:      kindFromExt(ext),
			timestamp: ts,
			size:      info.Size(),
			year:      ts.Format("2006"),
			month:     ts.Format("01"),
		})
		return nil
	})
	check(err)
}

// fileHash returns the SHA-256 hex digest of the file at path
func fileHash(path string) string {

	// open the file
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	// hash the contents
	hash := sha256.New()
	_, err = io.Copy(hash, file)
	check(err)
	return hex.EncodeToString(hash.Sum(nil))
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
