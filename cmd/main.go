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
		fmt.Println("Checking filenames with AI.")
		total := len(organised)
		failures := 0
		fmt.Print("0%")
		lastShown := 0
		for i, path := range organised {
			err := renameViaAI(url, model, apiKey, path)
			if err != nil {
				if isChatFailure(err) && !isFatalChat(err) {
					failures++
				} else {
					check(err)
				}
			}
			pct := ((i + 1) * 100) / total
			for next := lastShown + 10; next <= pct && next < 100; next += 10 {
				fmt.Printf("  %d%%", next)
				lastShown = next
			}
		}
		fmt.Println("  100%")
		if failures > 0 {
			fmt.Println("AI failures:", failures)
		}
	}

	fmt.Println()
	fmt.Println("Done.")
	fmt.Println()
	fmt.Println()
}
