package ai

import (
	"context"
	"fmt"
	"os"
)

// NewProvider is a factory function that returns a Provider based on the requested name.
func NewProvider(ctx context.Context, providerName string, modelName string) (Provider, error) {
	switch providerName {
	case "gemini", "":
		return NewGeminiClient(ctx, modelName)
	case "openai":
		return NewOpenAIClient(modelName)
	case "claude":
		return NewClaudeClient(modelName)
	case "ollama":
		return NewOllamaClient(modelName), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}
}

// GetDefaultProviderName returns the default provider based on environment variables or flags.
func GetDefaultProviderName(localFlag bool) string {
	if localFlag {
		return "ollama"
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		return "openai"
	}
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return "claude"
	}
	return "gemini"
}
