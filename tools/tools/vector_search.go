package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type VectorSearcher struct {
	store        *VectorStore
	openAIKey    string
	openAIClient *openai.Client
}

func NewVectorSearcher(store *VectorStore, openAIKey string) *VectorSearcher {
	return &VectorSearcher{
		store:        store,
		openAIKey:    openAIKey,
		openAIClient: openai.NewClient(openAIKey),
	}
}

type VectorSearchResponse struct {
	Chunks   []CodeChunk       `json:"chunks"`
	Summary  string            `json:"summary"`
	Metadata map[string]string `json:"metadata"`
}

func (vs *VectorSearcher) Search(query, repository, branch string) (*VectorSearchResponse, error) {
	// Get embedding for query
	embedding, err := vs.getEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %v", err)
	}

	// Get vector search results
	results, err := vs.store.Search(embedding, repository, branch, 10)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %v", err)
	}

	// Filter out .git files and select most relevant files
	var filteredResults []CodeChunk
	for _, result := range results {
		// Skip .git files and other metadata
		if strings.Contains(result.FilePath, ".git/") ||
			strings.HasPrefix(result.FilePath, ".git") ||
			strings.Contains(result.FilePath, "var/folders") {
			continue
		}

		// Skip empty or irrelevant content
		if len(strings.TrimSpace(result.Content)) < 10 {
			continue
		}

		filteredResults = append(filteredResults, result)
	}

	// Take top 3 most relevant results
	var topResults []CodeChunk
	if len(filteredResults) > 3 {
		topResults = filteredResults[:3]
	} else {
		topResults = filteredResults
	}

	// Generate summary using OpenAI with better prompt
	summary, err := vs.generateSummary(topResults, query)
	if err != nil {
		return nil, fmt.Errorf("summary generation failed: %v", err)
	}

	return &VectorSearchResponse{
		Chunks:  topResults,
		Summary: summary,
		Metadata: map[string]string{
			"repository": repository,
			"branch":     branch,
			"query":      query,
		},
	}, nil
}

func (vs *VectorSearcher) generateSummary(results []CodeChunk, query string) (string, error) {
	var contextBuilder strings.Builder
	for _, result := range results {
		contextBuilder.WriteString(fmt.Sprintf("File: %s\n```%s\n%s\n```\n\n",
			result.FilePath,
			result.Language,
			result.Content))
	}

	completion, err := vs.openAIClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: "system",
					Content: `You are a technical expert analyzing an e-commerce project. 
Focus on explaining the main features, architecture, and technologies used in the project. 
Provide specific details about the implementation and functionality.`,
				},
				{
					Role: "user",
					Content: fmt.Sprintf(`Analyze this e-commerce project and answer the query: %s

Project Context:
%s

Provide a detailed technical summary focusing on:
1. Main features and functionality
2. Technology stack and architecture
3. Key implementations
4. Notable patterns or practices used

Make the response specific to e-commerce functionality when possible.`,
						query, contextBuilder.String()),
				},
			},
			Temperature: 0.7,
			MaxTokens:   1000,
		},
	)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	return completion.Choices[0].Message.Content, nil
}

func (vs *VectorSearcher) getEmbedding(text string) ([]float32, error) {
	resp, err := vs.openAIClient.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	// Convert []float64 to []float32
	embedding := make([]float32, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}
