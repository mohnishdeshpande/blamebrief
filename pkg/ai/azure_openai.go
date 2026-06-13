package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// AzureOpenAIClient interfaces with Azure OpenAI API.
type AzureOpenAIClient struct {
	endpoint       string
	apiKey         string
	deploymentName string
}

// NewAzureOpenAIClient initializes a new Azure OpenAI client.
func NewAzureOpenAIClient(deploymentName string) (*AzureOpenAIClient, error) {
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("missing environment variable: AZURE_OPENAI_ENDPOINT")
	}

	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing environment variable: AZURE_OPENAI_API_KEY")
	}

	if deploymentName == "" {
		deploymentName = os.Getenv("AZURE_OPENAI_DEPLOYMENT_NAME")
		if deploymentName == "" {
			deploymentName = "gpt-4o" // Default expectation
		}
	}

	return &AzureOpenAIClient{
		endpoint:       endpoint,
		apiKey:         apiKey,
		deploymentName: deploymentName,
	}, nil
}

// ModelName returns the deployment name used by the client.
func (a *AzureOpenAIClient) ModelName() string {
	return a.deploymentName
}

// GenerateBrief generates the brief markdown explanation using Azure OpenAI.
func (a *AzureOpenAIClient) GenerateBrief(ctx context.Context, systemInstruction, prompt string) (string, error) {
	// Azure OpenAI endpoint format: https://{resource}.openai.azure.com/openai/deployments/{deployment}/chat/completions?api-version=2024-02-15-preview
	apiVersion := "2024-02-15-preview"
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", a.endpoint, a.deploymentName, apiVersion)

	reqBody := openAIRequest{
		Messages: []openAIMessage{
			{Role: "system", Content: systemInstruction},
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal azure openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request to Azure OpenAI: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", a.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to reach Azure OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	var resBody openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		return "", fmt.Errorf("failed to decode Azure OpenAI response JSON: %w", err)
	}

	if resBody.Error != nil {
		return "", fmt.Errorf("azure openai API error: %s", resBody.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("azure openai API returned status: %s", resp.Status)
	}

	if len(resBody.Choices) == 0 {
		return "", fmt.Errorf("received empty choices from Azure OpenAI API")
	}

	return resBody.Choices[0].Message.Content, nil
}
