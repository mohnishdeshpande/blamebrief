package ai

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

// GeminiClient wraps the official unified Google Gen AI Go SDK client.
type GeminiClient struct {
	client    *genai.Client
	modelName string
}

// NewGeminiClient initializes a new Gemini client.
// It auto-detects configuration:
// 1. If GCP_PROJECT is set, it initializes Vertex AI (BackendEnterprise).
// 2. Otherwise, it initializes Google AI Studio (BackendGeminiAPI) requiring GEMINI_API_KEY.
func NewGeminiClient(ctx context.Context, modelName string) (*GeminiClient, error) {
	if modelName == "" {
		modelName = "gemini-3.5-flash" // Best default for high-speed, long-context in 2026
	}

	var config *genai.ClientConfig

	// Check environment to determine which backend to use
	gcpProject := os.Getenv("GCP_PROJECT")
	gcpLocation := os.Getenv("GCP_LOCATION")
	if gcpLocation == "" {
		gcpLocation = "us-central1"
	}

	if gcpProject != "" {
		// Use Vertex AI
		config = &genai.ClientConfig{
			Project:  gcpProject,
			Location: gcpLocation,
			Backend:  genai.BackendEnterprise,
		}
	} else {
		// Use Google AI Studio
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("missing environment credentials: set GEMINI_API_KEY for Google AI Studio, or GCP_PROJECT (and optionally GCP_LOCATION) for Vertex AI")
		}
		config = &genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
		}
	}

	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create unified GenAI client: %w", err)
	}

	return &GeminiClient{
		client:    client,
		modelName: modelName,
	}, nil
}

// ModelName returns the resolved model name being used by the client.
func (g *GeminiClient) ModelName() string {
	return g.modelName
}

// GenerateBrief generates the brief markdown explanation using Gemini.
func (g *GeminiClient) GenerateBrief(ctx context.Context, systemInstruction, prompt string) (string, error) {
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: systemInstruction},
			},
		},
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.modelName, genai.Text(prompt), config)
	if err != nil {
		return "", fmt.Errorf("failed to generate brief from Gemini: %w", err)
	}

	text := resp.Text()
	if text == "" {
		return "", fmt.Errorf("received empty response text from Gemini API")
	}

	return text, nil
}
