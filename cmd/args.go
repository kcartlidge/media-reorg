package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// options holds the parsed CLI settings.
type options struct {
	Action string
	Folder string
	URL    string
	Model  string
	APIKey string
}

// parseArgs prints usage and returns validated options from named flags.
func parseArgs() options {
	fmt.Println("Usage:")
	fmt.Println("  media-reorg -action rearrange -folder <folder>")
	fmt.Println("  media-reorg -action rename -folder <folder> -api <url> -model <model> -api-key <api-key>")

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	action := fs.String("action", "", "")
	folder := fs.String("folder", "", "")
	api := fs.String("api", "", "")
	model := fs.String("model", "", "")
	apiKey := fs.String("api-key", "", "")

	check(fs.Parse(os.Args[1:]))
	if fs.NArg() > 0 {
		check(fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " ")))
	}

	opts := options{
		Action: strings.TrimSpace(*action),
		Folder: strings.TrimSpace(*folder),
		URL:    strings.TrimSpace(*api),
		Model:  strings.TrimSpace(*model),
		APIKey: strings.TrimSpace(*apiKey),
	}

	if opts.Action != "rearrange" && opts.Action != "rename" {
		check(fmt.Errorf("expected -action rearrange or rename"))
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

	switch opts.Action {
	case "rearrange":
		if hasAPI || hasModel || hasKey {
			check(fmt.Errorf("rearrange does not take -api, -model, or -api-key"))
		}
	case "rename":
		if !hasAPI || !hasModel {
			check(fmt.Errorf("rename requires -api and -model"))
		}
		opts.URL = strings.TrimRight(opts.URL, "/")
		if !strings.HasSuffix(opts.URL, "/v1") {
			opts.URL += "/v1"
		}
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
