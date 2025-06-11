package service

import (
	"fmt"

	"mcpserver/internal/models"
	"mcpserver/internal/storage"
)

type MCPServerService struct {
	pineconeStore  *storage.PineconeStore
	openaiClient   *storage.OpenAIClient
	vectorSearch   *VectorSearchService
	repoIndexer    *RepoIndexerService
}

func NewMCPServerService(pineconeStore *storage.PineconeStore, openaiClient *storage.OpenAIClient) *MCPServerService {
	return &MCPServerService{
		pineconeStore: pineconeStore,
		openaiClient:  openaiClient,
		vectorSearch:  NewVectorSearchService(pineconeStore, openaiClient),
		repoIndexer:   NewRepoIndexerService(pineconeStore, openaiClient),
	}
}

func (mcp *MCPServerService) GetServerInfo() *models.ServerInfo {
	return &models.ServerInfo{
		Name:    "Vector Search MCP",
		Version: "1.0.0",
		Capabilities: []string{
			"vector_search",
			"code_search",
			"repository_search",
			"repository_indexing",
		},
		Endpoints: map[string]string{
			"vector_search":    "/vector-search",
			"index_repository": "/index-repository",
			"health":           "/health",
		},
	}
}

func (mcp *MCPServerService) HandleCursorAction(action string, data map[string]interface{}) (interface{}, error) {
	switch action {
	case "connect":
		return map[string]string{"status": "connected"}, nil
	case "search":
		// Convert data to search request
		req := &models.SearchRequest{
			Query:      data["query"].(string),
			Repository: data["repository"].(string),
		}
		if branch, ok := data["branch"].(string); ok {
			req.Branch = branch
		}
		if limit, ok := data["limit"].(int); ok {
			req.Limit = limit
		}

		return mcp.vectorSearch.Search(req)
	default:
		return nil, fmt.Errorf("unknown cursor action: %s", action)
	}
}

func (mcp *MCPServerService) HandleChat(message, repository string, context map[string]interface{}) (interface{}, error) {
	// First, search for relevant code using vector search
	searchRequest := &models.SearchRequest{
		Query:      message,
		Repository: repository,
		Limit:      5,
	}

	searchResult, err := mcp.vectorSearch.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %v", err)
	}

	// Format response with search results
	return map[string]interface{}{
		"message":     "Here are some relevant code snippets I found:",
		"codeContext": searchResult,
	}, nil
}

func (mcp *MCPServerService) ConfigureGitHub(repository, token string) error {
	// Store GitHub configuration (implement secure storage)
	// This is a placeholder for actual implementation
	return nil
}