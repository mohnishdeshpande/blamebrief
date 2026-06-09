package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HistoryResult holds the extracted git history and the current HEAD commit hash.
type HistoryResult struct {
	RawLog      string
	CommitHash  string
	RepoRoot    string
	RelPath     string
	LineContent string
}

// ExtractHistory extracts the git log of a specific line range for a file.
// It resolves the file's repository root, gets the current HEAD commit, and runs 'git log -L'.
func ExtractHistory(filePath string, startLine, endLine int) (*HistoryResult, error) {
	// 1. Resolve absolute path of the file
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
	}

	// Verify file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", absPath)
	}

	// 2. Read the lines from the file
	lineContent, err := readLineRange(absPath, startLine, endLine)
	if err != nil {
		return nil, fmt.Errorf("failed to read line range from file: %w", err)
	}

	fileDir := filepath.Dir(absPath)

	// 3. Find git repository root
	repoRoot, err := getGitRepoRoot(fileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find git repository: %w", err)
	}

	// 4. Get relative path from repo root to the file
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find relative path from %s to %s: %w", repoRoot, absPath, err)
	}

	// Normalize windows separators just in case, git always uses forward slashes
	relPath = filepath.ToSlash(relPath)

	// 5. Get HEAD commit hash
	headHash, err := getHeadCommitHash(repoRoot)
	if err != nil {
		// Non-fatal if repo has no commits yet, but we'll report it
		return nil, fmt.Errorf("failed to get HEAD commit hash (is repository initialized with commits?): %w", err)
	}

	// 6. Run git log -L <start>,<end>:<relative_path>
	rawLog, err := runGitLogL(repoRoot, relPath, startLine, endLine)
	if err != nil {
		return nil, fmt.Errorf("failed to extract git line history: %w", err)
	}

	return &HistoryResult{
		RawLog:      rawLog,
		CommitHash:  headHash,
		RepoRoot:    repoRoot,
		RelPath:     relPath,
		LineContent: lineContent,
	}, nil
}

// readLineRange reads the lines between startLine and endLine (1-indexed, inclusive) from a file.
func readLineRange(filePath string, startLine, endLine int) (string, error) {
	if startLine <= 0 || endLine <= 0 || startLine > endLine {
		return "", fmt.Errorf("invalid line range: %d-%d", startLine, endLine)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lines []string
	currentLine := 1

	// We use a simple scanner
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if currentLine >= startLine && currentLine <= endLine {
			lines = append(lines, scanner.Text())
		}
		if currentLine > endLine {
			break
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if len(lines) == 0 {
		return "", fmt.Errorf("no lines found in range %d-%d", startLine, endLine)
	}

	return strings.Join(lines, "\n"), nil
}

// getGitRepoRoot runs 'git rev-parse --show-toplevel' in the specified directory.
func getGitRepoRoot(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", strings.TrimSpace(stderr.String()), err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// getHeadCommitHash runs 'git rev-parse HEAD' in the specified repository root.
func getHeadCommitHash(repoRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoRoot
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", strings.TrimSpace(stderr.String()), err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// runGitLogL runs 'git log -L <start>,<end>:<file>' in the repository root.
func runGitLogL(repoRoot, relPath string, startLine, endLine int) (string, error) {
	// git log -L requires a format like: -L 10,20:file.go
	rangeArg := fmt.Sprintf("-L %d,%d:%s", startLine, endLine, relPath)
	
	// We'll run: git log -L <start>,<end>:<file> --no-merges (or standard, let's keep it simple to allow merge commits if they have valuable info)
	cmd := exec.Command("git", "log", rangeArg)
	cmd.Dir = repoRoot
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Some git versions might fail if line numbers exceed file bounds. Give a clear explanation.
		stdErrStr := stderr.String()
		if strings.Contains(stdErrStr, "bounds") || strings.Contains(stdErrStr, "invalid") {
			return "", fmt.Errorf("invalid line range %d-%d for file %s. Git error: %s", startLine, endLine, relPath, strings.TrimSpace(stdErrStr))
		}
		return "", fmt.Errorf("%s: %s", strings.TrimSpace(stdErrStr), err)
	}

	return stdout.String(), nil
}

// CommitInfo holds parsed, structured information about a single git commit.
type CommitInfo struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	Date    string `json:"date"`
	Message string `json:"message"`
	Diff    string `json:"diff"`
}

// ParseRawLog parses the raw 'git log -L' output into structured CommitInfo items.
func ParseRawLog(rawLog string) []CommitInfo {
	var commits []CommitInfo
	lines := strings.Split(rawLog, "\n")
	
	var currentCommit *CommitInfo
	var inDiff bool
	var msgLines []string
	var diffLines []string

	for _, line := range lines {
		// Detect new commit boundary
		if strings.HasPrefix(line, "commit ") {
			// Save the previously assembled commit
			if currentCommit != nil {
				currentCommit.Message = strings.TrimSpace(strings.Join(msgLines, "\n"))
				currentCommit.Diff = strings.Join(diffLines, "\n")
				commits = append(commits, *currentCommit)
			}

			hash := strings.TrimPrefix(line, "commit ")
			currentCommit = &CommitInfo{Hash: hash}
			inDiff = false
			msgLines = nil
			diffLines = nil
			continue
		}

		if currentCommit == nil {
			continue
		}

		if strings.HasPrefix(line, "Author: ") {
			currentCommit.Author = strings.TrimSpace(strings.TrimPrefix(line, "Author: "))
			continue
		}

		if strings.HasPrefix(line, "Date: ") {
			currentCommit.Date = strings.TrimSpace(strings.TrimPrefix(line, "Date: "))
			continue
		}

		// Diff starts with "diff --git"
		if strings.HasPrefix(line, "diff --git ") {
			inDiff = true
			diffLines = append(diffLines, line)
			continue
		}

		if inDiff {
			diffLines = append(diffLines, line)
		} else {
			// Capture commit message text (omitting empty lines inside/outside indentation)
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" {
				msgLines = append(msgLines, trimmedLine)
			}
		}
	}

	// Capture the final commit in the list
	if currentCommit != nil {
		currentCommit.Message = strings.TrimSpace(strings.Join(msgLines, "\n"))
		currentCommit.Diff = strings.Join(diffLines, "\n")
		commits = append(commits, *currentCommit)
	}

	return commits
}
