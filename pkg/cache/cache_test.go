package cache

import (
	"os"
	"testing"
)

func TestCacheSetGetClear(t *testing.T) {
	// Create a temp directory for the cache
	tempDir, err := os.MkdirTemp("", "blamebrief-test-cache")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	c := &Cache{cacheDir: tempDir}

	hash := GenerateHash("testfile.go", 10, 20, "abcdef123456", "content")
	if hash == "" {
		t.Fatalf("hash was empty")
	}

	// Verify cache miss
	val, err := c.Get(hash)
	if err != nil {
		t.Fatalf("unexpected error on Get: %v", err)
	}
	if val != "" {
		t.Errorf("expected empty string on cache miss, got %q", val)
	}

	// Save to cache
	testBrief := "# Test Brief Content"
	err = c.Set(hash, testBrief)
	if err != nil {
		t.Fatalf("failed to Set cache: %v", err)
	}

	// Verify cache hit
	val, err = c.Get(hash)
	if err != nil {
		t.Fatalf("unexpected error on Get: %v", err)
	}
	if val != testBrief {
		t.Errorf("expected cached brief to be %q, got %q", testBrief, val)
	}

	// Clear cache
	err = c.Clear()
	if err != nil {
		t.Fatalf("failed to Clear cache: %v", err)
	}

	// Verify cache miss again after clearing
	val, err = c.Get(hash)
	if err != nil {
		t.Fatalf("unexpected error on Get after clear: %v", err)
	}
	if val != "" {
		t.Errorf("expected empty string after clearing cache, got %q", val)
	}
}
