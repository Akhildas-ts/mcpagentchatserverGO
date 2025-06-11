package storage

import (
	"context"
	"fmt"
	"log"
	"strings"

	"mcpserver/internal/models"

	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	log.Printf("OPENAI_API_KEY present: %v", apiKey != "")
	
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
	}
}

func (oc *OpenAIClient) GetEmbedding(text string) ([]float32, error) {
	resp, err := oc.client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Model: openai.AdaEmbeddingV2,
			Input: []string{text},
		},
	)
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

func (oc *OpenAIClient) GenerateEnhancedSummary(chunks []map[string]interface{}, query string) (string, error) {
	if oc.client == nil {
		return "", fmt.Errorf("OpenAI client not initialized")
	}

	var contextBuilder strings.Builder
	for _, chunk := range chunks {
		contextBuilder.WriteString(fmt.Sprintf("File: %s\nContent:\n%s\n\n",
			chunk["filePath"],
			chunk["content"]))
	}

	completion, err := oc.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: "system",
					Content: `You are a technical expert. Provide ONLY direct answers to queries about code repositories.
- Answer the specific question asked
- Be concise and to the point
- Do not include additional context unless specifically asked
- If the answer is found, just state it directly`,
				},
				{
					Role: "user",
					Content: fmt.Sprintf(`Question: %s

Code Context:
%s

Provide only the direct answer to the question.`,
						query, contextBuilder.String()),
				},
			},
			Temperature: 0.3, // Lower temperature for more focused responses
			MaxTokens:   200, // Reduced tokens for shorter responses
		},
	)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	return completion.Choices[0].Message.Content, nil
}

func (oc *OpenAIClient) GenerateSummary(chunks []models.CodeChunk, query string) (string, error) {
	var contextBuilder strings.Builder
	for _, chunk := range chunks {
		contextBuilder.WriteString(fmt.Sprintf("File: %s\n```%s\n%s\n```\n\n",
			chunk.FilePath,
			chunk.Language,
			chunk.Content))
	}

	completion, err := oc.client.CreateChatCompletion(
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