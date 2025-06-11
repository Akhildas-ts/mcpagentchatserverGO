package models

// CodeChunk represents a chunk of code with metadata
type CodeChunk struct {
	Content    string    `json:"content"`
	FilePath   string    `json:"filePath"`
	Repository string    `json:"repository"`
	Branch     string    `json:"branch"`
	Language   string    `json:"language"`
	Embedding  []float32 `json:"embedding"`
}

// SearchRequest represents a vector search request
type SearchRequest struct {
	Query      string `json:"query"`
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Limit      int    `json:"limit"`
}

// SearchResponse represents a vector search response
type SearchResponse struct {
	Chunks []CodeChunk `json:"chunks"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Content    string `json:"content"`
	FilePath   string `json:"filePath"`
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Language   string `json:"language"`
	Embedding  any    `json:"embedding"`
}

// ServerInfo represents MCP server information
type ServerInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Capabilities []string          `json:"capabilities"`
	Endpoints    map[string]string `json:"endpoints"`
}

// IndexRepositoryRequest represents a repository indexing request
type IndexRepositoryRequest struct {
	RepoURL string `json:"repoUrl"`
	Branch  string `json:"branch"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Message    string                 `json:"message"`
	Repository string                 `json:"repository"`
	Context    map[string]interface{} `json:"context"`
}

// CursorRequest represents a cursor connection request
type CursorRequest struct {
	CursorID string                 `json:"cursorId"`
	Action   string                 `json:"action"`
	Data     map[string]interface{} `json:"data"`
}

// GitHubConfigRequest represents GitHub configuration request
type GitHubConfigRequest struct {
	Repository string `json:"repository"`
	Token      string `json:"token"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}