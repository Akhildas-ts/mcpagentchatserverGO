package tools

import (
	"os"
	"testing"
)

func TestNewRepoIndexer(t *testing.T) {
	tests := []struct {
		name         string
		pineconeKey  string
		pineconeEnv  string
		indexName    string
		pineconeHost string
		openAIKey    string
		expectError  bool
	}{
		{
			name:         "Valid configuration",
			pineconeKey:  "test-pinecone-key",
			pineconeEnv:  "test-env",
			indexName:    "test-index",
			pineconeHost: "test-host",
			openAIKey:    "test-openai-key",
			expectError:  false,
		},
		{
			name:         "Missing Pinecone key",
			pineconeKey:  "",
			pineconeEnv:  "test-env",
			indexName:    "test-index",
			pineconeHost: "test-host",
			openAIKey:    "test-openai-key",
			expectError:  true,
		},
		{
			name:         "Missing OpenAI key",
			pineconeKey:  "test-pinecone-key",
			pineconeEnv:  "test-env",
			indexName:    "test-index",
			pineconeHost: "test-host",
			openAIKey:    "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexer, err := NewRepoIndexer(tt.pineconeKey, tt.pineconeEnv, tt.indexName, tt.pineconeHost, tt.openAIKey)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if indexer == nil {
					t.Error("Expected non-nil indexer")
				}
			}
		})
	}
}

func TestRepoIndexer_IndexRepository(t *testing.T) {
	// Setup test environment
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("PINECONE_API_KEY", "test-key")
	os.Setenv("PINECONE_ENVIRONMENT", "test-env")
	os.Setenv("PINECONE_INDEX_NAME", "test-index")
	os.Setenv("PINECONE_HOST", "test-host")

	indexer, err := NewRepoIndexer(
		os.Getenv("PINECONE_API_KEY"),
		os.Getenv("PINECONE_ENVIRONMENT"),
		os.Getenv("PINECONE_INDEX_NAME"),
		os.Getenv("PINECONE_HOST"),
		os.Getenv("OPENAI_API_KEY"),
	)
	if err != nil {
		t.Fatalf("Failed to create repo indexer: %v", err)
	}

	tests := []struct {
		name        string
		repoURL     string
		branch      string
		expectError bool
	}{
		{
			name:        "Valid public repository",
			repoURL:     "https://github.com/golang/example.git",
			branch:      "master",
			expectError: false,
		},
		{
			name:        "Invalid repository URL",
			repoURL:     "invalid-url",
			branch:      "main",
			expectError: true,
		},
		{
			name:        "Non-existent branch",
			repoURL:     "https://github.com/golang/example.git",
			branch:      "non-existent-branch",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := indexer.IndexRepository(tt.repoURL, tt.branch)
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

func TestRepoIndexer_ProcessFile(t *testing.T) {
	indexer := createTestRepoIndexer()

	tests := []struct {
		name        string
		content     string
		filePath    string
		repoURL     string
		branch      string
		expectError bool
	}{
		{
			name:        "Valid Go file",
			content:     "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}",
			filePath:    "main.go",
			repoURL:     "https://github.com/golang/example.git",
			branch:      "master",
			expectError: false,
		},
		{
			name:        "Empty file",
			content:     "",
			filePath:    "empty.go",
			repoURL:     "https://github.com/golang/example.git",
			branch:      "master",
			expectError: true,
		},
		{
			name:        "Binary file",
			content:     string([]byte{0x00, 0x01, 0x02, 0x03}),
			filePath:    "binary.bin",
			repoURL:     "https://github.com/golang/example.git",
			branch:      "master",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := indexer.processFile(tt.content, tt.filePath, tt.repoURL, tt.branch)
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

func TestRepoIndexer_Execute(t *testing.T) {
	indexer := createTestRepoIndexer()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid parameters",
			params: map[string]interface{}{
				"repoUrl": "https://github.com/golang/example.git",
				"branch":  "master",
			},
			expectError: false,
		},
		{
			name: "Missing repoUrl",
			params: map[string]interface{}{
				"branch": "main",
			},
			expectError: true,
		},
		{
			name:        "Nil parameters",
			params:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := indexer.Execute(tt.params)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected non-nil result")
				}
			}
		})
	}
}
