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
	fmt.Println("  media-reorg <folder> [<url> <model> [<api-key>]]")

	// get the source folder and optional LLM config
	folder := sourceFolder()
	url, model, apiKey, useLLM := llmConfig()

	fmt.Println()
	fmt.Println("Folder:", folder)
	if useLLM {
		fmt.Println()
		fmt.Println("Using an Open AI compatible API to rename files:")
		fmt.Println("     URL =", url)
		fmt.Println("   Model =", model)
		if apiKey != "" {
			fmt.Println(" API Key =", obfuscateKey(apiKey))
		}

		// basic check that the API and model are valid
		models, err := listModels(url, apiKey)
		check(err)
		found := false
		for _, id := range models {
			if id == model {
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
			check(fmt.Errorf("model not found: %s", model))
		}
	}

	// scan the source folder
	fmt.Println()
	fmt.Println("Scanning.")
	fmt.Println()
	scan(folder)

	// move files into year/month folders
	fmt.Println()
	fmt.Println("Organising.")
	organised := processEntries(folder)

	// remove empty and junk-only folders
	fmt.Println()
	fmt.Println("Cleaning up.")
	clearup(folder)

	// optional AI pass over filenames in the main dated folders
	if useLLM {
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
			did, err := renameViaAI(url, model, apiKey, path)
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
