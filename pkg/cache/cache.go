package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// Cache handles saving and retrieving generated briefs to/from local storage.
type Cache struct {
	cacheDir string
}

// NewCache initializes and returns a Cache instance, ensuring the cache directory exists.
func NewCache() (*Cache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	cacheDir := filepath.Join(home, ".blamebrief", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory at %s: %w", cacheDir, err)
	}

	return &Cache{cacheDir: cacheDir}, nil
}

// GenerateHash generates a unique SHA256 string for the file, line range, HEAD commit, and line content.
func GenerateHash(relPath string, startLine, endLine int, commitHash, lineContent string) string {
	hasher := sha256.New()
	hasher.Write([]byte(relPath))
	hasher.Write([]byte(fmt.Sprintf(":%d-%d:", startLine, endLine)))
	hasher.Write([]byte(commitHash))
	hasher.Write([]byte(":" + lineContent))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Get retrieves a cached brief if it exists.
func (c *Cache) Get(hash string) (string, error) {
	cacheFile := filepath.Join(c.cacheDir, hash+".md")
	
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return "", nil // Cache miss, no error
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", fmt.Errorf("failed to read cache file: %w", err)
	}

	return string(data), nil
}

// Set saves a brief to the cache.
func (c *Cache) Set(hash string, brief string) error {
	cacheFile := filepath.Join(c.cacheDir, hash+".md")
	
	err := os.WriteFile(cacheFile, []byte(brief), 0644)
	if err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Clear removes all files in the cache directory.
func (c *Cache) Clear() error {
	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		err := os.Remove(filepath.Join(c.cacheDir, entry.Name()))
		if err != nil {
			return fmt.Errorf("failed to remove cache entry %s: %w", entry.Name(), err)
		}
	}
	return nil
}
