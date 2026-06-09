package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaClient_GenerateBrief(t *testing.T) {
	// Set up local mock Ollama server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request details
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/generate" {
			t.Errorf("expected /api/generate path, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req ollamaRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			t.Fatalf("failed to decode mock request: %v", err)
		}

		if req.Model != "test-gemma" {
			t.Errorf("expected model 'test-gemma', got %q", req.Model)
		}
		if req.Prompt != "test prompt" {
			t.Errorf("expected prompt 'test prompt', got %q", req.Prompt)
		}
		if req.System != "test system instruction" {
			t.Errorf("expected system 'test system instruction', got %q", req.System)
		}

		// Send mock response
		resp := ollamaResponse{
			Response: "This is a simulated Software Archaeologist brief from Gemma.",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Initialize OllamaClient pointing to our mock server
	client := &OllamaClient{
		baseURL:   mockServer.URL,
		modelName: "test-gemma",
	}

	brief, err := client.GenerateBrief(context.Background(), "test system instruction", "test prompt")
	if err != nil {
		t.Fatalf("GenerateBrief failed: %v", err)
	}

	expectedBrief := "This is a simulated Software Archaeologist brief from Gemma."
	if brief != expectedBrief {
		t.Errorf("expected brief %q, got %q", expectedBrief, brief)
	}
}
