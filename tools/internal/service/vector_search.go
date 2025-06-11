package service

import (
	"fmt"
	"log"

	"mcpserver/internal/models"
	"mcpserver/internal/storage"
)

type VectorSearchService struct {
	pineconeStore *storage.PineconeStore
	openaiClient  *storage.OpenAIClient
}

func NewVectorSearchService(pineconeStore *storage.PineconeStore, openaiClient *storage.OpenAIClient) *VectorSearchService {
	return &VectorSearchService{
		pineconeStore: pineconeStore,
		openaiClient:  openaiClient,
	}
}

func (vs *VectorSearchService) Search(req *models.SearchRequest) (*models.SearchResponse, error) {
	// Get query embedding
	embedding, err := vs.openaiClient.GetEmbedding(req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %v", err)
	}

	// Set default branch if not provided
	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	// Set default limit if not provided
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	// Search vector store with branch filter
	chunks, err := vs.pineconeStore.Search(embedding, req.Repository, branch, limit)
	if err != nil {
		return nil, fmt.Errorf("vector store search failed: %v", err)
	}

	log.Printf("Found %d chunks from vector store\n", len(chunks))

	return &models.SearchResponse{
		Chunks: chunks,
	}, nil
}

func (vs *VectorSearchService) SearchWithSummary(req *models.SearchRequest) (map[string]interface{}, error) {
	// Perform regular search
	searchResponse, err := vs.Search(req)
	if err != nil {
		return nil, err
	}

	// Convert chunks to format expected by OpenAI service
	chunks := make([]map[string]interface{}, len(searchResponse.Chunks))
	for i, chunk := range searchResponse.Chunks {
		chunks[i] = map[string]interface{}{
			"content":    chunk.Content,
			"filePath":   chunk.FilePath,
			"repository": chunk.Repository,
			"branch":     chunk.Branch,
			"language":   chunk.Language,
		}
	}

	// Generate summary using OpenAI
	summary, err := vs.openaiClient.GenerateEnhancedSummary(chunks, req.Query)
	if err != nil {
		return nil, fmt.Errorf("summary generation failed: %v", err)
	}

	// Return response with summary
	return map[string]interface{}{
		"summary": summary,
	}, nil
}