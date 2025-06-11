package mocks

import (
	"errors"

	"mcpserver/internal/models"
)

// MockPineconeStore provides a mock implementation of the Pinecone store
type MockPineconeStore struct {
	err error
}

func NewMockPineconeStore() *MockPineconeStore {
	return &MockPineconeStore{}
}

func (m *MockPineconeStore) SetError(err error) {
	m.err = err
}

func (m *MockPineconeStore) Store(chunk models.CodeChunk) error {
	if m.err != nil {
		return m.err
	}
	if chunk.Content == "" {
		return errors.New("empty content")
	}
	return nil
}

func (m *MockPineconeStore) Search(query []float32, repository, branch string, limit int) ([]models.CodeChunk, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(query) == 0 {
		return nil, errors.New("empty query")
	}
	if limit <= 0 {
		return nil, errors.New("invalid limit")
	}

	return []models.CodeChunk{
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

// MockOpenAIClient provides a mock implementation of the OpenAI client
type MockOpenAIClient struct {
	err error
}

func NewMockOpenAIClient() *MockOpenAIClient {
	return &MockOpenAIClient{}
}

func (m *MockOpenAIClient) SetError(err error) {
	m.err = err
}

func (m *MockOpenAIClient) GetEmbedding(text string) ([]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	if text == "" {
		return nil, errors.New("empty text")
	}
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *MockOpenAIClient) GenerateEnhancedSummary(chunks []map[string]interface{}, query string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "Test summary response", nil
}

func (m *MockOpenAIClient) GenerateSummary(chunks []models.CodeChunk, query string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "Test summary response", nil
}