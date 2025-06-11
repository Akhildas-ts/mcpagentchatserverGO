package storage

import (
	"context"
	"fmt"
	"log"
	"strings"

	"mcpserver/internal/models"

	"github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

type PineconeStore struct {
	client      *pinecone.Client
	indexName   string
	environment string
	hostUrl     string
}

func NewPineconeStore(apiKey, environment, indexName, hostUrl string) (*PineconeStore, error) {
	log.Printf("PINECONE_API_KEY present: %v", apiKey != "")
	log.Printf("PINECONE_ENVIRONMENT present: %v", environment != "")
	log.Printf("PINECONE_INDEX_NAME present: %v", indexName != "")
	log.Printf("PINECONE_HOST present: %v", hostUrl != "")

	// Initialize the client
	client, err := pinecone.NewClient(
		pinecone.NewClientParams{
			ApiKey: apiKey,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pinecone client: %w", err)
	}

	return &PineconeStore{
		client:      client,
		indexName:   indexName,
		environment: environment,
		hostUrl:     hostUrl,
	}, nil
}

func (ps *PineconeStore) Search(query []float32, repository string, branch string, limit int) ([]models.CodeChunk, error) {
	ctx := context.Background()

	fmt.Printf("Searching for repository: %s, branch: %s with limit: %d\n", repository, branch, limit)

	index, err := ps.client.Index(pinecone.NewIndexConnParams{
		Host: ps.hostUrl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	fmt.Printf("Connected to Pinecone index: %s at %s\n", ps.indexName, ps.hostUrl)

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

	var prioritizedResults []models.CodeChunk
	var otherResults []models.CodeChunk

	// Parse results
	for i, match := range queryResp.Matches {
		if match == nil || match.Vector == nil || match.Vector.Metadata == nil {
			fmt.Printf("Match %d is nil or has nil vector/metadata\n", i)
			continue
		}
		metadata := match.Vector.Metadata.AsMap()
		fmt.Printf("Match %d - ID: %s, Score: %f\n", i, match.Vector.Id, match.Score)

		chunk := models.CodeChunk{
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

func (ps *PineconeStore) Store(chunk models.CodeChunk) error {
	ctx := context.Background()

	fmt.Printf("Storing chunk for repository: %s, filepath: %s\n", chunk.Repository, chunk.FilePath)

	index, err := ps.client.Index(pinecone.NewIndexConnParams{
		Host: ps.hostUrl,
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