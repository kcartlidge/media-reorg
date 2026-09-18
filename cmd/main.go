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

	opts := parseArgs()

	fmt.Println()
	fmt.Println("Folder:", opts.Folder)
	if opts.UseLLM {
		fmt.Println()
		fmt.Println("Using an Open AI compatible API to rename files:")
		fmt.Println("     URL =", opts.URL)
		fmt.Println("   Model =", opts.Model)
		if opts.APIKey != "" {
			fmt.Println(" API Key =", obfuscateKey(opts.APIKey))
		}

		// basic check that the API and model are valid
		models, err := listModels(opts.URL, opts.APIKey)
		check(err)
		found := false
		for _, id := range models {
			if id == opts.Model {
				found = true
				break
			}
		}
		if !found {
			fmt.Println()
			fmt.Println("Available models:")
			for _, id := range models {
				fmt.Println(" -", id)
			}
			check(fmt.Errorf("model not found: %s", opts.Model))
		}
	}

	// scan the source folder
	fmt.Println()
	fmt.Println("Scanning.")
	fmt.Println()
	scan(opts.Folder)

	// move files into year/month folders
	fmt.Println()
	fmt.Println("Organising.")
	organised := processEntries(opts.Folder)

	// remove empty and junk-only folders
	fmt.Println()
	fmt.Println("Cleaning up.")
	clearup(opts.Folder)

	// optional AI pass over filenames in the main dated folders
	if opts.UseLLM {
		fmt.Println()
		total := len(organised)
		fmt.Println("Image files:", total)
		renamed := 0
		failures := 0
		fmt.Print("0")
		early := []int{10, 25, 50}
		earlyIdx := 0
		nextHundred := 100
		for i, path := range organised {
			did, err := renameViaAI(opts.URL, opts.Model, opts.APIKey, path)
			if err != nil {
				if isChatFailure(err) && !isFatalChat(err) {
					failures++
					parkAIFailure(path)
				} else {
					check(err)
				}
			} else if did {
				renamed++
			}
			n := i + 1
			for earlyIdx < len(early) && n >= early[earlyIdx] {
				fmt.Printf("  %d", early[earlyIdx])
				earlyIdx++
			}
			for n >= nextHundred {
				fmt.Printf("  %d", nextHundred)
				nextHundred += 100
			}
		}
		fmt.Println()
		fmt.Println()
		fmt.Println("Image files:", total)
		fmt.Println("Renamed via AI:", renamed)
		fmt.Println("AI failures:", failures)
	}

	fmt.Println()
	fmt.Println("Done.")
	fmt.Println()
	fmt.Println()
}
