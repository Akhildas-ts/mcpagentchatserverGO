package tools

import (
	"testing"
)

func TestNewVectorStore(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		env         string
		indexName   string
		hostUrl     string
		expectError bool
	}{
		{
			name:        "Valid configuration",
			apiKey:      "test-api-key",
			env:         "test-env",
			indexName:   "test-index",
			hostUrl:     "test-host",
			expectError: false,
		},
		{
			name:        "Missing API key",
			apiKey:      "",
			env:         "test-env",
			indexName:   "test-index",
			hostUrl:     "test-host",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewVectorStore(tt.apiKey, tt.env, tt.indexName, tt.hostUrl)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if store == nil {
					t.Error("Expected non-nil store")
				}
			}
		})
	}
}

func TestVectorStore_Store(t *testing.T) {
	store := NewMockVectorStore()

	tests := []struct {
		name        string
		chunk       CodeChunk
		expectError bool
	}{
		{
			name: "Valid chunk",
			chunk: CodeChunk{
				Content:    "test content",
				FilePath:   "test/path.go",
				Repository: "test/repo",
				Branch:     "main",
				Language:   "Go",
				Embedding:  []float32{0.1, 0.2, 0.3},
			},
			expectError: false,
		},
		{
			name: "Empty content",
			chunk: CodeChunk{
				Content:    "",
				FilePath:   "test/path.go",
				Repository: "test/repo",
				Branch:     "main",
				Language:   "Go",
				Embedding:  []float32{0.1, 0.2, 0.3},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Store(tt.chunk)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestVectorStore_Search(t *testing.T) {
	store := NewMockVectorStore()

	tests := []struct {
		name        string
		query       []float32
		repository  string
		branch      string
		limit       int
		expectError bool
	}{
		{
			name:        "Valid search",
			query:       []float32{0.1, 0.2, 0.3},
			repository:  "test/repo",
			branch:      "main",
			limit:       10,
			expectError: false,
		},
		{
			name:        "Empty query",
			query:       []float32{},
			repository:  "test/repo",
			branch:      "main",
			limit:       10,
			expectError: true,
		},
		{
			name:        "Invalid limit",
			query:       []float32{0.1, 0.2, 0.3},
			repository:  "test/repo",
			branch:      "main",
			limit:       -1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := store.Search(tt.query, tt.repository, tt.branch, tt.limit)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if results == nil {
					t.Error("Expected non-nil results")
				}
			}
		})
	}
}

// Mock implementations for testing
type mockPineconeClient struct {
	// Add mock methods as needed
}

type mockOpenAIClient struct {
	// Add mock methods as needed
}
