package main

import (
	"fmt"
)

func main() {
	fmt.Println()
	fmt.Println()
	fmt.Println("MEDIA-REORG")
	fmt.Println("K Cartlidge, 2026")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  media-reorg <folder>")

	// get the source folder
	folder := sourceFolder()

	fmt.Println()
	fmt.Println("Folder:", folder)

	// scan the source folder
	fmt.Println()
	fmt.Println("Scanning.")
	fmt.Println()
	scan(folder)

	// move files into year/month folders
	fmt.Println()
	fmt.Println("Organising.")
	processEntries(folder)

	// remove empty and junk-only folders
	fmt.Println()
	fmt.Println("Cleaning up.")
	clearup(folder)
	fmt.Println()
	fmt.Println()
}
