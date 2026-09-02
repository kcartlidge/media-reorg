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

	folder := sourceFolder()

	fmt.Println()
	fmt.Println("Folder:", folder)
	fmt.Println()
	fmt.Println()
}
