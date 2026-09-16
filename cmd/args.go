package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// sourceFolder returns the path to the source folder
func sourceFolder() string {

	// require folder, optionally url+model[+api-key]
	n := len(os.Args)
	if n != 2 && n != 4 && n != 5 {
		check(fmt.Errorf("expected <folder> [<url> <model> [<api-key>]]"))
	}

	// resolve to an absolute path
	path, err := filepath.Abs(os.Args[1])
	check(err)

	// ensure it exists and is a folder
	info, err := os.Stat(path)
	check(err)
	if !info.IsDir() {
		check(fmt.Errorf("not a folder: %s", path))
	}

	return path
}

// llmConfig returns optional OpenAI-compatible API settings from the CLI.
// ok is false when no LLM args were given.
func llmConfig() (url, model, apiKey string, ok bool) {

	// no LLM config provided
	if len(os.Args) < 4 {
		return "", "", "", false
	}

	url = strings.TrimSpace(os.Args[2])
	model = strings.TrimSpace(os.Args[3])
	if url == "" {
		check(fmt.Errorf("expected <url>"))
	}
	if model == "" {
		check(fmt.Errorf("expected <model>"))
	}

	// OpenAI-compatible APIs expect a /v1 base path
	url = strings.TrimRight(url, "/")
	if !strings.HasSuffix(url, "/v1") {
		url += "/v1"
	}

	// optional API key
	if len(os.Args) == 5 {
		apiKey = strings.TrimSpace(os.Args[4])
		if apiKey == "" {
			check(fmt.Errorf("expected <api-key>"))
		}
	}

	return url, model, apiKey, true
}

// obfuscateKey shows the start and end of a key with the middle hidden
func obfuscateKey(key string) string {
	if len(key) < 6 {
		return "..."
	}
	return key[:3] + "..." + key[len(key)-3:]
}
