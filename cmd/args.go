package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// options holds the parsed CLI settings for the current combined run.
type options struct {
	Folder string
	URL    string
	Model  string
	APIKey string
	UseLLM bool
}

// parseArgs prints usage and returns validated options from named flags.
func parseArgs() options {
	fmt.Println("Usage:")
	fmt.Println("  media-reorg -folder <folder> [-api <url> -model <model> [-api-key <api-key>]]")

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	folder := fs.String("folder", "", "")
	api := fs.String("api", "", "")
	model := fs.String("model", "", "")
	apiKey := fs.String("api-key", "", "")

	check(fs.Parse(os.Args[1:]))
	if fs.NArg() > 0 {
		check(fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " ")))
	}

	opts := options{
		Folder: strings.TrimSpace(*folder),
		URL:    strings.TrimSpace(*api),
		Model:  strings.TrimSpace(*model),
		APIKey: strings.TrimSpace(*apiKey),
	}

	if opts.Folder == "" {
		check(fmt.Errorf("expected -folder"))
	}
	path, err := filepath.Abs(opts.Folder)
	check(err)
	info, err := os.Stat(path)
	check(err)
	if !info.IsDir() {
		check(fmt.Errorf("not a folder: %s", path))
	}
	opts.Folder = path

	hasAPI := opts.URL != ""
	hasModel := opts.Model != ""
	hasKey := opts.APIKey != ""
	if hasAPI != hasModel {
		check(fmt.Errorf("expected -api and -model together"))
	}
	if hasKey && !hasAPI {
		check(fmt.Errorf("expected -api-key only with -api and -model"))
	}
	if hasAPI {
		opts.URL = strings.TrimRight(opts.URL, "/")
		if !strings.HasSuffix(opts.URL, "/v1") {
			opts.URL += "/v1"
		}
		opts.UseLLM = true
	}

	return opts
}

// obfuscateKey shows the start and end of a key with the middle hidden
func obfuscateKey(key string) string {
	if len(key) < 6 {
		return "..."
	}
	return key[:3] + "..." + key[len(key)-3:]
}
