package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

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

// llmModelsResponse is the OpenAI-compatible GET /models payload
type llmModelsResponse struct {
	Data []llmModel `json:"data"`
}

// llmModel is one entry from the models list
type llmModel struct {
	ID string `json:"id"`
}
