package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// BedrockClaudeClient interfaces with AWS Bedrock Claude models using the AWS CLI.
type BedrockClaudeClient struct {
	modelID string
}

type bedrockClaudeRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	MaxTokens        int             `json:"max_tokens"`
	System           string          `json:"system"`
	Messages         []claudeMessage `json:"messages"`
}

type bedrockClaudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// NewBedrockClaudeClient initializes a new Bedrock Claude client.
func NewBedrockClaudeClient(modelID string) *BedrockClaudeClient {
	if modelID == "" {
		modelID = "anthropic.claude-3-5-sonnet-20240620-v1:0"
	}
	return &BedrockClaudeClient{
		modelID: modelID,
	}
}

// ModelName returns the model ID used by the client.
func (b *BedrockClaudeClient) ModelName() string {
	return b.modelID
}

// GenerateBrief generates the brief markdown explanation using AWS Bedrock Claude.
func (b *BedrockClaudeClient) GenerateBrief(ctx context.Context, systemInstruction, prompt string) (string, error) {
	reqBody := bedrockClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
		System:           systemInstruction,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal bedrock request: %w", err)
	}

	// We use the AWS CLI to avoid complex SigV4 signing logic and heavy SDK dependencies.
	// This is consistent with the project's strategy of using native tools like git.
	cmd := exec.CommandContext(ctx, "aws", "bedrock-runtime", "invoke-model",
		"--model-id", b.modelID,
		"--body", string(payload),
		"/dev/stdout", // Output to stdout
	)

	// Ensure AWS region is set if provided in environment
	if region := os.Getenv("AWS_REGION"); region != "" {
		cmd.Env = append(os.Environ(), "AWS_DEFAULT_REGION="+region)
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("aws bedrock-runtime failed (%w): %s", err, string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute aws cli: %w", err)
	}

	var resBody bedrockClaudeResponse
	if err := json.NewDecoder(bytes.NewReader(output)).Decode(&resBody); err != nil {
		return "", fmt.Errorf("failed to decode bedrock response JSON: %w", err)
	}

	if len(resBody.Content) == 0 {
		return "", fmt.Errorf("received empty content from bedrock")
	}

	return resBody.Content[0].Text, nil
}
