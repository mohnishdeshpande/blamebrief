package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// OllamaClient interfaces with local Ollama server.
type OllamaClient struct {
	baseURL   string
	modelName string
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

// NewOllamaClient creates a client pointing to OLLAMA_HOST or default localhost.
func NewOllamaClient(modelName string) *OllamaClient {
	baseURL := os.Getenv("OLLAMA_HOST")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if modelName == "" {
		modelName = "gemma4" // Recommended default local Google model
	}
	return &OllamaClient{
		baseURL:   baseURL,
		modelName: modelName,
	}
}

// ModelName returns the resolved model name being used by the client.
func (o *OllamaClient) ModelName() string {
	return o.modelName
}

// GenerateBrief queries the local Ollama server to get a code-evolution brief.
func (o *OllamaClient) GenerateBrief(ctx context.Context, systemInstruction, prompt string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", o.baseURL)

	reqBody := ollamaRequest{
		Model:  o.modelName,
		Prompt: prompt,
		System: systemInstruction,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request to Ollama: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Set a generous timeout since local generation might be slow depending on the machine specs
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to reach Ollama at %s (please ensure Ollama is installed and running): %w", o.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned error status: %s", resp.Status)
	}

	var resBody ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		return "", fmt.Errorf("failed to decode Ollama response JSON: %w", err)
	}

	return resBody.Response, nil
}
