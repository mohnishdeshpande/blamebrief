package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"blamebrief/pkg/ai"
	"blamebrief/pkg/cache"
	"blamebrief/pkg/git"

	"github.com/spf13/cobra"
)

var (
	linesFlag      string
	detailFlag     string
	localFlag      bool
	providerFlag   string
	modelFlag      string
	clearCacheFlag bool
	rawFlag        bool
	jsonFlag       bool
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "blamebrief [file path]",
	Short: "BlameBrief is an AI-powered software archeology CLI tool that deciphers code history.",
	Long: `BlameBrief goes beyond 'git blame'. It extracts the deep chronological git history
and evolution of a specific code block, and pipes into an AI provider (Gemini, Claude, GPT, or Ollama) 
to analyze why the code was written this way, its intent, and hidden risks (Chesterton's Fence).`,
	Args: cobra.ExactArgs(1),
	RunE: runBlameBrief,
}

func init() {
	RootCmd.Flags().StringVarP(&linesFlag, "lines", "l", "", "Line range to analyze (e.g., '10-45' or '25'). If omitted, analyzes the whole file.")
	RootCmd.Flags().StringVarP(&detailFlag, "detail", "d", "high", "Level of detail for the report (high, medium, low)")
	RootCmd.Flags().BoolVar(&localFlag, "local", false, "Use local Ollama instance running Gemma (shorthand for --provider ollama)")
	RootCmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "AI provider to use (gemini, openai, claude, ollama)")
	RootCmd.Flags().StringVar(&modelFlag, "model", "", "Override default AI model name (e.g. 'gemini-1.5-pro', 'gpt-4o', 'claude-3-5-sonnet-20240620')")
	RootCmd.Flags().BoolVar(&clearCacheFlag, "clear-cache", false, "Clear the local blamebrief cache before running")
	RootCmd.Flags().BoolVar(&rawFlag, "raw", false, "Output the raw gathered Git history context package as Markdown (No AI)")
	RootCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output the structured Git history context package as JSON (No AI)")
}

func runBlameBrief(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// 1. Initialize cache
	c, err := cache.NewCache()
	if err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	// 2. Handle clear cache flag
	if clearCacheFlag {
		fmt.Println("[blamebrief] Clearing local cache...")
		if err := c.Clear(); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
		fmt.Println("[blamebrief] Cache cleared successfully.")
	}

	// 3. Count total lines for strict bounds verification
	totalLines, err := countLines(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// 4. Parse and validate line range
	var startLine, endLine int
	if linesFlag != "" {
		startLine, endLine, err = parseLines(linesFlag)
		if err != nil {
			return fmt.Errorf("failed to parse lines: %w", err)
		}
		if startLine > totalLines || endLine > totalLines {
			return fmt.Errorf("invalid range %d-%d: file %s only has %d lines", startLine, endLine, filePath, totalLines)
		}
	} else {
		// Default to entire file
		startLine = 1
		endLine = totalLines
		fmt.Printf("[blamebrief] Analyzing entire file (lines %d-%d)...\n", startLine, endLine)
	}

	// 5. Extract Git history and line contents
	fmt.Printf("[blamebrief] Digging up Git history for %s (lines %d-%d)...\n", filePath, startLine, endLine)
	history, err := git.ExtractHistory(filePath, startLine, endLine)
	if err != nil {
		return err
	}

	// 5a. Deterministic Output Modes (No AI / LLMs)
	if jsonFlag {
		parsedCommits := git.ParseRawLog(history.RawLog)
		jsonBytes, err := json.MarshalIndent(parsedCommits, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal history to JSON: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	if rawFlag {
		fmt.Printf("# BlameBrief Raw Context: %s (Lines %d-%d)\n\n", history.RelPath, startLine, endLine)
		fmt.Println("## Current Code Block")
		fmt.Printf("```\n%s\n```\n\n", history.LineContent)
		fmt.Println("## Chronological Git History")
		fmt.Println(history.RawLog)
		return nil
	}

	// 5b. Check cache unless clear-cache was used (avoid double checking)
	cacheHash := cache.GenerateHash(history.RelPath, startLine, endLine, history.CommitHash, history.LineContent)
	if !clearCacheFlag {
		cachedBrief, err := c.Get(cacheHash)
		if err == nil && cachedBrief != "" {
			fmt.Println("\n[blamebrief] Serving cached brief (instant search):")
			fmt.Println(cachedBrief)
			return nil
		}
	}

	// 6. Formulate Prompts
	systemInstruction := `You are a Software Archaeologist, a Senior Principal Engineer with decades of experience diagnosing legacy systems, deciphering intent, and mapping technological debt. Your specialty is understanding the "Chesterton's Fence" of codebases: why code was written the way it was, what past incidents occurred, what workarounds were implemented, and what risks are hidden.

Your goal is to output a clean, technical, and objective markdown summary ("Brief") of the evolution of the provided code range. Do not use corporate fluff, patronizing language, or generic filler. Speak with extreme competence and technical clarity.`

	prompt := fmt.Sprintf(`Analyze the historical evolution of the following code block.

FILE PATH: %s
LINE RANGE: %d-%d

=== CURRENT CODE BLOCK ===
%s

=== CHRONOLOGICAL GIT HISTORY (DIFF EVOLUTIONS) ===
%s

=== INSTRUCTIONS ===
1. Study the entire chronological evolution of this code block from the oldest commit to the newest.
2. Formulate a technical narrative explaining *why* this code block was modified at each stage.
3. Identify specific reasons for the current design:
   - Was it a bug fix? What was the bug?
   - Was it an architectural optimization?
   - Is it a "workaround" or temporary hotfix for an external system/incident?
4. Call out "The Wow Factors" if applicable:
   - Frustration/Sentiment level: Note if authors seemed stressed or rushed in commit messages (detect words like 'hotfix', 'temporary', 'please ignore', 'hack', 'workaround', etc.) and estimate a "Frustration Score" out of 10.
   - Technical Debt warnings: Explain what would break if a developer naively simplified or removed parts of this code (Chesterton's Fence warning).
5. Output format must be clean, structured Markdown. Start with a '# BlameBrief: %s (Lines %d-%d)' heading.
6. Provide sections exactly matching these headers:
   - **Executive Summary**: A 2-3 sentence overview of why this code looks the way it does today.
   - **Chronological Timeline**: Bullet points of key commits, their authors, dates, and the core change intent.
   - **Architectural Intent & Chesterton's Fence Warning**: What is the hidden purpose of this code? What does it protect against? What are the dependencies?
   - **Technical Debt & Sentiment Score**: If any frustration or quick-fixes are detected, call them out. Give a Frustration Score from 1 to 10 with explanation.

Detail Level: %s`,
		history.RelPath, startLine, endLine, history.LineContent, history.RawLog, history.RelPath, startLine, endLine, detailFlag)

	// 7. Execute AI Generation
	ctx := context.Background()

	// Determine provider
	if providerFlag == "" {
		providerFlag = ai.GetDefaultProviderName(localFlag)
	}

	p, err := ai.NewProvider(ctx, providerFlag, modelFlag)
	if err != nil {
		return err
	}

	fmt.Printf("[blamebrief] Querying %s (%s)... This may take a moment.\n", providerFlag, p.ModelName())
	brief, err := p.GenerateBrief(ctx, systemInstruction, prompt)
	if err != nil {
		return fmt.Errorf("%s analysis failed: %w", providerFlag, err)
	}

	// 8. Cache the result
	_ = c.Set(cacheHash, brief) // Ignore write failure, still output result

	// 9. Display Brief
	fmt.Println("\n" + brief)
	return nil
}

func parseLines(linesFlag string) (int, int, error) {
	linesFlag = strings.TrimSpace(linesFlag)
	if linesFlag == "" {
		return 0, 0, fmt.Errorf("lines flag is required if provided")
	}

	var start, end int
	var err error

	// Try splitting by hyphen or colon
	sep := "-"
	if strings.Contains(linesFlag, ":") {
		sep = ":"
	}

	parts := strings.Split(linesFlag, sep)
	if len(parts) == 1 {
		part0 := strings.TrimSpace(parts[0])
		start, err = strconv.Atoi(part0)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid line number: %s", part0)
		}
		end = start
	} else if len(parts) == 2 {
		part0 := strings.TrimSpace(parts[0])
		start, err = strconv.Atoi(part0)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid start line number: %s", part0)
		}
		part1 := strings.TrimSpace(parts[1])
		end, err = strconv.Atoi(part1)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid end line number: %s", part1)
		}
	} else {
		return 0, 0, fmt.Errorf("invalid lines format. Use 'start-end' (e.g. 10-45)")
	}

	if start <= 0 || end <= 0 || start > end {
		return 0, 0, fmt.Errorf("invalid range: lines must be positive and start line must be <= end line")
	}

	return start, end, nil
}

func countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}
