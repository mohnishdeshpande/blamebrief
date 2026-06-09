package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// ClaudeClient interfaces with Anthropic API.
type ClaudeClient struct {
	apiKey    string
	modelName string
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewClaudeClient initializes a new Claude client.
func NewClaudeClient(modelName string) (*ClaudeClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing environment variable: ANTHROPIC_API_KEY")
	}

	if modelName == "" {
		modelName = "claude-3-5-sonnet-20240620"
	}

	return &ClaudeClient{
		apiKey:    apiKey,
		modelName: modelName,
	}, nil
}

// ModelName returns the resolved model name being used by the client.
func (c *ClaudeClient) ModelName() string {
	return c.modelName
}

// GenerateBrief generates the brief markdown explanation using Claude.
func (c *ClaudeClient) GenerateBrief(ctx context.Context, systemInstruction, prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"

	reqBody := claudeRequest{
		Model:     c.modelName,
		MaxTokens: 4096,
		System:    systemInstruction,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claude request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request to Claude: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to reach Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	var resBody claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		return "", fmt.Errorf("failed to decode Claude response JSON: %w", err)
	}

	if resBody.Error != nil {
		return "", fmt.Errorf("claude API error (%s): %s", resBody.Error.Type, resBody.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("claude API returned status: %s", resp.Status)
	}

	if len(resBody.Content) == 0 {
		return "", fmt.Errorf("received empty content from Claude API")
	}

	return resBody.Content[0].Text, nil
}
