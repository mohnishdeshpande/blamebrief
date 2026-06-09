package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// OpenAIClient interfaces with OpenAI API.
type OpenAIClient struct {
	apiKey    string
	modelName string
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewOpenAIClient initializes a new OpenAI client.
func NewOpenAIClient(modelName string) (*OpenAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing environment variable: OPENAI_API_KEY")
	}

	if modelName == "" {
		modelName = "gpt-4o"
	}

	return &OpenAIClient{
		apiKey:    apiKey,
		modelName: modelName,
	}, nil
}

// ModelName returns the resolved model name being used by the client.
func (o *OpenAIClient) ModelName() string {
	return o.modelName
}

// GenerateBrief generates the brief markdown explanation using OpenAI.
func (o *OpenAIClient) GenerateBrief(ctx context.Context, systemInstruction, prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := openAIRequest{
		Model: o.modelName,
		Messages: []openAIMessage{
			{Role: "system", Content: systemInstruction},
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request to OpenAI: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to reach OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	var resBody openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		return "", fmt.Errorf("failed to decode OpenAI response JSON: %w", err)
	}

	if resBody.Error != nil {
		return "", fmt.Errorf("openai API error: %s", resBody.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai API returned status: %s", resp.Status)
	}

	if len(resBody.Choices) == 0 {
		return "", fmt.Errorf("received empty choices from OpenAI API")
	}

	return resBody.Choices[0].Message.Content, nil
}
