package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed prompts/*.txt
var promptFS embed.FS

// loadPrompt returns a trimmed prompt from the embedded prompts folder
func loadPrompt(name string) string {
	data, err := promptFS.ReadFile("prompts/" + name)
	check(err)
	return strings.TrimSpace(string(data))
}

// chatError marks a chat API failure; Fatal means stop the run
type chatError struct {
	Fatal bool
	err   error
}

func (e *chatError) Error() string { return e.err.Error() }
func (e *chatError) Unwrap() error { return e.err }

// listModels fetches available model ids from an OpenAI-compatible API
func listModels(baseURL, apiKey string) ([]string, error) {

	// always send a key; local servers often ignore it
	if apiKey == "" {
		apiKey = "n/a"
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/models"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models request failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload llmModelsResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(payload.Data))
	for _, m := range payload.Data {
		ids = append(ids, m.ID)
	}
	return ids, nil
}

// askChat sends a simple user prompt to an OpenAI-compatible chat API
func askChat(baseURL, apiKey, model, prompt string) (string, error) {
	return doChat(baseURL, apiKey, model, prompt)
}

// askChatImage sends a prompt plus an image to an OpenAI-compatible chat API
func askChatImage(baseURL, apiKey, model, prompt, imagePath string) (string, error) {

	data, err := os.ReadFile(imagePath)
	if err != nil {
		return "", err
	}

	mime := imageMIME(filepath.Ext(imagePath))
	url := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
	content := []llmContentPart{
		{Type: "text", Text: prompt},
		{Type: "image_url", ImageURL: &llmImageURL{URL: url}},
	}
	return doChat(baseURL, apiKey, model, content)
}

// doChat posts a chat completion with either text or multimodal content
func doChat(baseURL, apiKey, model string, content any) (string, error) {

	// always send a key; local servers often ignore it
	if apiKey == "" {
		apiKey = "n/a"
	}

	payload, err := json.Marshal(llmChatRequest{
		Model: model,
		Messages: []llmChatMessage{
			{Role: "user", Content: content},
		},
	})
	if err != nil {
		return "", &chatError{err: err}
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", &chatError{err: err}
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", &chatError{Fatal: true, err: fmt.Errorf("chat connectivity failed: %w", err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &chatError{err: err}
	}
	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(string(body))
		err := fmt.Errorf("chat request failed (%d): %s", resp.StatusCode, msg)
		if missingModelStatus(resp.StatusCode, msg) {
			return "", &chatError{Fatal: true, err: err}
		}
		return "", &chatError{err: err}
	}

	var result llmChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", &chatError{err: err}
	}
	if len(result.Choices) == 0 {
		return "", &chatError{err: fmt.Errorf("chat response had no choices")}
	}
	return result.Choices[0].Message.Content, nil
}

// missingModelStatus reports a missing/unknown model from status or body text
func missingModelStatus(status int, body string) bool {
	if status == http.StatusNotFound {
		return true
	}
	lower := strings.ToLower(body)
	return strings.Contains(lower, "model_not_found") ||
		strings.Contains(lower, "model not found") ||
		strings.Contains(lower, "invalid model") ||
		(strings.Contains(lower, "model") && strings.Contains(lower, "does not exist"))
}

// isFatalChat reports whether err is a chat failure that should stop the run
func isFatalChat(err error) bool {
	var ce *chatError
	return errors.As(err, &ce) && ce.Fatal
}

// isChatFailure reports whether err came from a chat request
func isChatFailure(err error) bool {
	var ce *chatError
	return errors.As(err, &ce)
}

// imageMIME returns a MIME type for a common image extension
func imageMIME(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".tif", ".tiff":
		return "image/tiff"
	case ".heic":
		return "image/heic"
	case ".heif":
		return "image/heif"
	case ".avif":
		return "image/avif"
	default:
		return "image/jpeg"
	}
}

// llmModelsResponse is the OpenAI-compatible GET /models payload
type llmModelsResponse struct {
	Data []llmModel `json:"data"`
}

// llmModel is one entry from the models list
type llmModel struct {
	ID string `json:"id"`
}

// llmChatRequest is the OpenAI-compatible POST /chat/completions payload
type llmChatRequest struct {
	Model    string           `json:"model"`
	Messages []llmChatMessage `json:"messages"`
}

// llmChatMessage is one chat message; Content may be a string or multimodal parts
type llmChatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// llmChatResponse is the OpenAI-compatible chat completion result
type llmChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// llmContentPart is one multimodal content item
type llmContentPart struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	ImageURL *llmImageURL `json:"image_url,omitempty"`
}

// llmImageURL holds a data-URL or remote image reference
type llmImageURL struct {
	URL string `json:"url"`
}
