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
	fmt.Println()

	// scan the source folder
	scan(folder)

	// move files into year/month folders
	processEntries(folder)
	fmt.Println()
	fmt.Println()
}
