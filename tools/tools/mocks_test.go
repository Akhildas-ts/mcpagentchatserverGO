package tools

import (
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
)

// Mock OpenAI Client
type MockOpenAIClient struct {
	err error
}

func NewMockOpenAIClient() *MockOpenAIClient {
	return &MockOpenAIClient{}
}

func (m *MockOpenAIClient) CreateEmbeddings(ctx context.Context, request openai.EmbeddingRequestConverter) (openai.EmbeddingResponse, error) {
	if m.err != nil {
		return openai.EmbeddingResponse{}, m.err
	}
	return openai.EmbeddingResponse{
		Data: []openai.Embedding{
			{
				Embedding: []float32{0.1, 0.2, 0.3},
			},
		},
	}, nil
}

func (m *MockOpenAIClient) CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	if m.err != nil {
		return openai.ChatCompletionResponse{}, m.err
	}
	return openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Content: "Test response",
				},
			},
		},
	}, nil
}

// Mock Vector Search Tool
type MockVectorSearchTool struct {
	err error
}

func NewMockVectorSearchTool() VectorSearchToolInterface {
	return &MockVectorSearchTool{}
}

func (m *MockVectorSearchTool) Execute(params map[string]interface{}) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}

	// Check required parameters
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, errors.New("query is required")
	}

	repository, ok := params["repository"].(string)
	if !ok || repository == "" {
		return nil, errors.New("repository is required")
	}

	// Return mock results
	return map[string]interface{}{
		"chunks": []map[string]interface{}{
			{
				"content":    "test content",
				"filePath":   "test/path.go",
				"repository": repository,
				"branch":     "main",
				"language":   "Go",
			},
		},
	}, nil
}

// Mock Vector Store
type MockVectorStore struct {
	err error
}

func NewMockVectorStore() VectorStoreInterface {
	return &MockVectorStore{}
}

func (m *MockVectorStore) Store(chunk CodeChunk) error {
	if m.err != nil {
		return m.err
	}
	if chunk.Content == "" {
		return errors.New("empty content")
	}
	return nil
}

func (m *MockVectorStore) Search(query []float32, repository, branch string, limit int) ([]CodeChunk, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(query) == 0 {
		return nil, errors.New("empty query")
	}
	if limit <= 0 {
		return nil, errors.New("invalid limit")
	}

	return []CodeChunk{
		{
			Content:    "test content",
			FilePath:   "test/path.go",
			Repository: repository,
			Branch:     branch,
			Language:   "Go",
			Embedding:  []float32{0.1, 0.2, 0.3},
		},
	}, nil
}

// Helper function to create a test RepoIndexer with mocks
func createTestRepoIndexer() *RepoIndexer {
	return &RepoIndexer{
		vectorStore:  NewMockVectorStore(),
		openAIClient: NewMockOpenAIClient(),
	}
}
