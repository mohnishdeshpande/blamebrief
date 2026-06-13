package ai

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CLIProvider handles LLM calls by delegating to an external CLI tool already authenticated in the environment.
type CLIProvider struct {
	binaryName string
	args       []string
}

// NewCopilotProvider creates a provider that uses the GitHub Copilot CLI or Codex binary.
func NewCopilotProvider() *CLIProvider {
	// If 'codex' is in path, use it, otherwise fallback to 'gh copilot'
	binary := "gh"
	args := []string{"copilot", "explain"}
	
	if path, err := exec.LookPath("codex"); err == nil && path != "" {
		binary = "codex"
		args = []string{} // Assuming 'codex [prompt]' works based on user description
	}
	
	return &CLIProvider{
		binaryName: binary,
		args:       args,
	}
}

// NewClaudeCodeProvider creates a provider that uses the Claude Code CLI.
func NewClaudeCodeProvider() *CLIProvider {
	return &CLIProvider{
		binaryName: "claude",
		args:       []string{"--non-interactive"},
	}
}

// ModelName returns the name of the binary being used.
func (c *CLIProvider) ModelName() string {
	return c.binaryName
}

// GenerateBrief delegates the prompt to the CLI tool.
func (c *CLIProvider) GenerateBrief(ctx context.Context, systemInstruction, prompt string) (string, error) {
	fullPrompt := fmt.Sprintf("%s\n\n%s", systemInstruction, prompt)
	
	var cmdArgs []string
	cmdArgs = append(cmdArgs, c.args...)
	
	// Specific handling for different CLIs
	switch c.binaryName {
	case "gh":
		// gh copilot explain [prompt]
		cmdArgs = append(cmdArgs, fullPrompt)
	case "claude":
		// claude [prompt] --non-interactive
		cmdArgs = append(cmdArgs, fullPrompt)
	default:
		cmdArgs = append(cmdArgs, fullPrompt)
	}

	cmd := exec.CommandContext(ctx, c.binaryName, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("CLI delegation failed (%s): %w\nOutput: %s", c.binaryName, err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}
