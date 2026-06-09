package ai

import "context"

// Provider defines the common interface for all AI clients.
type Provider interface {
	// GenerateBrief sends the system instructions and user prompt to the AI and returns the result.
	GenerateBrief(ctx context.Context, systemInstruction, prompt string) (string, error)

	// ModelName returns the name of the model being used.
	ModelName() string
}
