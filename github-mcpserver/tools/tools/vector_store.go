package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/pinecone-io/go-pinecone/pinecone"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/types/known/structpb"
)

type VectorStore struct {
	client      *pinecone.Client
	indexName   string
	environment string
	hostUrl     string
}

type CodeChunk struct {
	Content    string    `json:"content"`
	FilePath   string    `json:"filePath"`
	Repository string    `json:"repository"`
	Branch     string    `json:"branch"`
	Language   string    `json:"language"`
	Embedding  []float32 `json:"embedding"`
}

func NewVectorStore(apiKey, environment, indexName, hostUrl string) (*VectorStore, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("pinecone API key is required")
	}

	// Initialize the client
	client, err := pinecone.NewClient(
		pinecone.NewClientParams{
			ApiKey: apiKey,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pinecone client: %w", err)
	}

	return &VectorStore{
		client:      client,
		indexName:   indexName,
		environment: environment,
		hostUrl:     hostUrl,
	}, nil
}

func (vs *VectorStore) Search(query []float32, repository string, branch string, limit int) ([]CodeChunk, error) {
	ctx := context.Background()

	fmt.Printf("Searching for repository: %s, branch: %s with limit: %d\n", repository, branch, limit)

	index, err := vs.client.Index(pinecone.NewIndexConnParams{
		Host: vs.hostUrl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	fmt.Printf("Connected to Pinecone index: %s at %s\n", vs.indexName, vs.hostUrl)

	// Convert repository and branch filter to structpb
	filterStruct, err := structpb.NewStruct(map[string]interface{}{
		"repository": repository,
		"branch":     branch,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create filter: %w", err)
	}

	fmt.Printf("Using filter: repository=%s, branch=%s\n", repository, branch)

	// Perform query
	queryResp, err := index.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:          query,
		TopK:            uint32(limit),
		MetadataFilter:  filterStruct,
		IncludeMetadata: true,
	})
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	fmt.Printf("Query response received, matches count: %d\n", len(queryResp.Matches))

	// Add boost for important files
	importantFiles := []string{
		"main.go",
		"README.md",
		"go.mod",
		"handlers/",
		"models/",
		"routes/",
		"controllers/",
		"services/",
	}

	var prioritizedResults []CodeChunk
	var otherResults []CodeChunk

	// Parse results
	for i, match := range queryResp.Matches {
		if match == nil || match.Vector == nil || match.Vector.Metadata == nil {
			fmt.Printf("Match %d is nil or has nil vector/metadata\n", i)
			continue
		}
		metadata := match.Vector.Metadata.AsMap()
		fmt.Printf("Match %d - ID: %s, Score: %f\n", i, match.Vector.Id, match.Score)

		chunk := CodeChunk{
			Content:    metadata["content"].(string),
			FilePath:   metadata["filePath"].(string),
			Repository: metadata["repository"].(string),
			Branch:     metadata["branch"].(string),
			Language:   metadata["language"].(string),
		}

		// Prioritize important files
		isImportant := false
		for _, importantFile := range importantFiles {
			if strings.Contains(chunk.FilePath, importantFile) {
				prioritizedResults = append(prioritizedResults, chunk)
				isImportant = true
				break
			}
		}
		if !isImportant {
			otherResults = append(otherResults, chunk)
		}
	}

	// Combine results with priority
	allResults := append(prioritizedResults, otherResults...)

	fmt.Printf("Returning %d chunks\n", len(allResults))
	return allResults, nil
}

func (vs *VectorStore) Store(chunk CodeChunk) error {
	ctx := context.Background()

	fmt.Printf("Storing chunk for repository: %s, filepath: %s\n", chunk.Repository, chunk.FilePath)

	index, err := vs.client.Index(pinecone.NewIndexConnParams{
		Host: vs.hostUrl,
	})
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	// Convert metadata to structpb
	metadata, err := structpb.NewStruct(map[string]interface{}{
		"content":    chunk.Content,
		"filePath":   chunk.FilePath,
		"repository": chunk.Repository,
		"branch":     chunk.Branch,
		"language":   chunk.Language,
	})
	if err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}

	// Create a unique ID for the vector
	vectorId := fmt.Sprintf("%s-%s", chunk.Repository, chunk.FilePath)
	if len(vectorId) > 100 {
		// Ensure ID is not too long for Pinecone
		vectorId = vectorId[:100]
	}

	// Create vector
	vectors := []*pinecone.Vector{
		{
			Id:       vectorId,
			Values:   chunk.Embedding,
			Metadata: metadata,
		},
	}

	// Perform upsert
	resp, err := index.UpsertVectors(ctx, vectors)
	if err != nil {
		return fmt.Errorf("failed to store chunk: %w", err)
	}

	fmt.Printf("Successfully stored chunk. Upserted: %v\n", resp)

	return nil
}

type VectorSearchTool struct {
	vectorStore  *VectorStore
	openAIClient *openai.Client
}

type SearchRequest struct {
	Query      string `json:"query"`
	Repository string `json:"repository"`
	Limit      int    `json:"limit"`
}

type SearchResponse struct {
	Chunks []CodeChunk `json:"chunks"`
}

func NewVectorSearchTool(pineconeAPIKey, pineconeEnv, pineconeIndex, pineconeHost, openAIKey string) (*VectorSearchTool, error) {
	vectorStore, err := NewVectorStore(pineconeAPIKey, pineconeEnv, pineconeIndex, pineconeHost)
	if err != nil {
		return nil, err
	}

	openAIClient := openai.NewClient(openAIKey)

	return &VectorSearchTool{
		vectorStore:  vectorStore,
		openAIClient: openAIClient,
	}, nil
}

func (t *VectorSearchTool) Execute(params map[string]interface{}) (interface{}, error) {
	// Extract parameters with type checking
	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter must be a string")
	}

	repository, ok := params["repository"].(string)
	if !ok {
		return nil, fmt.Errorf("repository parameter must be a string")
	}

	branch, ok := params["branch"].(string)
	if !ok {
		branch = "main" // default branch if not provided
	}

	// Get query embedding
	embedding, err := t.getQueryEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %v", err)
	}

	// Search vector store with branch filter
	chunks, err := t.vectorStore.Search(embedding, repository, branch, 10)
	if err != nil {
		return nil, fmt.Errorf("vector store search failed: %v", err)
	}

	fmt.Printf("Found %d chunks from vector store\n", len(chunks))

	// Convert chunks to response format
	searchResults := make([]map[string]interface{}, len(chunks))
	for i, chunk := range chunks {
		searchResults[i] = map[string]interface{}{
			"content":    chunk.Content,
			"filePath":   chunk.FilePath,
			"repository": chunk.Repository,
			"branch":     chunk.Branch,
			"language":   chunk.Language,
			"embedding":  nil,
		}
		fmt.Printf("Processed chunk %d: %s\n", i, chunk.FilePath)
	}

	return map[string]interface{}{
		"chunks": searchResults,
	}, nil
}

func (t *VectorSearchTool) getQueryEmbedding(query string) ([]float32, error) {
	resp, err := t.openAIClient.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Model: openai.AdaEmbeddingV2,
			Input: []string{query},
		},
	)
	if err != nil {
		return nil, err
	}

	// Convert []float64 to []float32
	embedding := make([]float32, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

type SearchResult struct {
	Content    string `json:"content"`
	FilePath   string `json:"filePath"`
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Language   string `json:"language"`
}
