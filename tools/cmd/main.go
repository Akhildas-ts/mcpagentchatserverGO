package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"mcpserver/tools"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type DefaultToolHandler struct {
	tools map[string]func(map[string]interface{}) (interface{}, error)
}

func NewDefaultToolHandler() *DefaultToolHandler {
	return &DefaultToolHandler{
		tools: make(map[string]func(map[string]interface{}) (interface{}, error)),
	}
}

func (h *DefaultToolHandler) RegisterTool(name string, handler func(map[string]interface{}) (interface{}, error)) {
	h.tools[name] = handler
}

func (h *DefaultToolHandler) ExecuteTool(name string, params map[string]interface{}) (interface{}, error) {
	if handler, ok := h.tools[name]; ok {
		return handler(params)
	}
	return nil, fmt.Errorf("tool %s not found", name)
}

type MCPServer struct {
	ToolHandler  ToolHandler
	openAIClient *openai.Client
}

type ToolHandler interface {
	RegisterTool(name string, handler func(map[string]interface{}) (interface{}, error))
	ExecuteTool(name string, params map[string]interface{}) (interface{}, error)
}

type ServerInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Capabilities []string          `json:"capabilities"`
	Endpoints    map[string]string `json:"endpoints"`
}

func sendResponse(w http.ResponseWriter, success bool, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"data":    data,
		"message": message,
	})
}

func (s *MCPServer) SetupVectorSearch() error {
	log.Printf("PINECONE_API_KEY present: %v", os.Getenv("PINECONE_API_KEY") != "")
	log.Printf("PINECONE_ENVIRONMENT present: %v", os.Getenv("PINECONE_ENVIRONMENT") != "")
	log.Printf("PINECONE_INDEX_NAME present: %v", os.Getenv("PINECONE_INDEX_NAME") != "")
	log.Printf("PINECONE_HOST present: %v", os.Getenv("PINECONE_HOST") != "")
	log.Printf("OPENAI_API_KEY present: %v", os.Getenv("OPENAI_API_KEY") != "")

	vectorSearchTool, err := tools.NewVectorSearchTool(
		os.Getenv("PINECONE_API_KEY"),
		os.Getenv("PINECONE_ENVIRONMENT"),
		os.Getenv("PINECONE_INDEX_NAME"),
		os.Getenv("PINECONE_HOST"),
		os.Getenv("OPENAI_API_KEY"),
	)
	if err != nil {
		fmt.Println("error from setup vector search", err)
		return err
	}

	s.ToolHandler.RegisterTool("vectorSearch", vectorSearchTool.Execute)
	return nil
}

func (s *MCPServer) HandleVectorSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Query      string `json:"query"`
		Repository string `json:"repository"`
		Branch     string `json:"branch"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, false, nil, "Invalid request format")
		return
	}

	if req.Query == "" || req.Repository == "" {
		sendResponse(w, false, nil, "Query and repository are required")
		return
	}

	// If branch is not provided, use default
	if req.Branch == "" {
		req.Branch = "main"
	}

	// Prepare parameters for vector search
	params := map[string]interface{}{
		"query":      req.Query,
		"repository": req.Repository,
		"branch":     req.Branch,
	}

	// Execute vector search
	results, err := s.ToolHandler.ExecuteTool("vectorSearch", params)
	if err != nil {
		sendResponse(w, false, nil, fmt.Sprintf("Search failed: %v", err))
		return
	}

	// First, assert the map
	resultsMap, ok := results.(map[string]interface{})
	if !ok {
		sendResponse(w, false, nil, "Invalid results format")
		return
	}

	// Then get the chunks array
	chunks, ok := resultsMap["chunks"].([]map[string]interface{})
	if !ok {
		sendResponse(w, false, nil, "Invalid chunks format")
		return
	}

	// Now you can process the chunks
	// Generate summary using OpenAI
	summary, err := s.generateEnhancedSummary(chunks, req.Query)
	if err != nil {
		sendResponse(w, false, nil, fmt.Sprintf("Summary generation failed: %v", err))
		return
	}

	// Send only the summary in the response
	response := map[string]interface{}{
		"summary": summary,
	}

	sendResponse(w, true, response, "")
}

func (s *MCPServer) generateEnhancedSummary(chunks []map[string]interface{}, query string) (string, error) {
	if s.openAIClient == nil {
		return "", fmt.Errorf("OpenAI client not initialized")
	}

	var contextBuilder strings.Builder
	for _, chunk := range chunks {
		contextBuilder.WriteString(fmt.Sprintf("File: %s\nContent:\n%s\n\n",
			chunk["filePath"],
			chunk["content"]))
	}

	completion, err := s.openAIClient.CreateChatCompletion(
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

func formatResults(results []SearchResult) string {
	var sb strings.Builder
	for _, result := range results {
		sb.WriteString(fmt.Sprintf("File: %s\n%s\n\n", result.FilePath, result.Content))
	}
	return sb.String()
}

type SearchResult struct {
	Content    string `json:"content"`
	FilePath   string `json:"filePath"`
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Language   string `json:"language"`
	Embedding  any    `json:"embedding"`
}

func (s *MCPServer) HandleCursorConnection(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for cursor connections
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		CursorID string                 `json:"cursorId"`
		Action   string                 `json:"action"`
		Data     map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, false, nil, "Invalid cursor request format")
		return
	}

	// Handle different cursor actions
	switch req.Action {
	case "connect":
		// Handle cursor connection
		sendResponse(w, true, map[string]string{"status": "connected"}, "Cursor connected successfully")
	case "search":
		// Reuse your vector search functionality
		result, err := s.ToolHandler.ExecuteTool("vectorSearch", req.Data)
		if err != nil {
			sendResponse(w, false, nil, fmt.Sprintf("Search failed: %v", err))
			return
		}
		sendResponse(w, true, result, "")
	default:
		sendResponse(w, false, nil, "Unknown cursor action")
	}
}

func (s *MCPServer) HandleMCPRegistration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	serverInfo := ServerInfo{
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

	sendResponse(w, true, serverInfo, "")
}

func (s *MCPServer) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	sendResponse(w, true, map[string]string{"status": "healthy"}, "")
}

func (s *MCPServer) HandleChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Message    string                 `json:"message"`
		Repository string                 `json:"repository"`
		Context    map[string]interface{} `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, false, nil, "Invalid chat request format")
		return
	}

	// First, search for relevant code using vector search
	searchResult, err := s.ToolHandler.ExecuteTool("vectorSearch", map[string]interface{}{
		"query":      req.Message,
		"repository": req.Repository,
		"limit":      5,
	})
	if err != nil {
		sendResponse(w, false, nil, fmt.Sprintf("Search failed: %v", err))
		return
	}

	// Format response with search results
	response := map[string]interface{}{
		"message":     "Here are some relevant code snippets I found:",
		"codeContext": searchResult,
	}

	sendResponse(w, true, response, "")
}

func (s *MCPServer) HandleGitHubConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var req struct {
		Repository string `json:"repository"`
		Token      string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, false, nil, "Invalid configuration request")
		return
	}

	// Store GitHub configuration (implement secure storage)
	// ...

	sendResponse(w, true, map[string]string{"status": "configured"}, "GitHub repository configured")
}

func (s *MCPServer) SetupRepoIndexer() error {
	repoIndexer, err := tools.NewRepoIndexer(
		os.Getenv("PINECONE_API_KEY"),
		os.Getenv("PINECONE_ENVIRONMENT"),
		os.Getenv("PINECONE_INDEX_NAME"),
		os.Getenv("PINECONE_HOST"),
		os.Getenv("OPENAI_API_KEY"),
	)
	if err != nil {
		fmt.Println("error from setup repo indexer:", err)
		return err
	}

	s.ToolHandler.RegisterTool("indexRepository", repoIndexer.Execute)
	return nil
}

func (s *MCPServer) HandleRepositoryIndexing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		RepoURL string `json:"repoUrl"`
		Branch  string `json:"branch"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, false, nil, "Invalid request format")
		return
	}

	params := map[string]interface{}{
		"repoUrl": req.RepoURL,
		"branch":  req.Branch,
	}

	result, err := s.ToolHandler.ExecuteTool("indexRepository", params)
	if err != nil {
		sendResponse(w, false, nil, fmt.Sprintf("Repository indexing failed: %v", err))
		return
	}

	sendResponse(w, true, result, "Repository indexed successfully")
}

func (s *MCPServer) authenticateRequest(r *http.Request) bool {
	token := r.Header.Get("X-MCP-Token")
	// Implement your authentication logic
	return token == os.Getenv("MCP_SECRET_TOKEN")
}

func (s *MCPServer) logRequest(handler string, req interface{}, err error) {
	if err != nil {
		log.Printf("[ERROR] %s: %v - Request: %+v", handler, err, req)
	}
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Create a new MCP server with the default tool handler
	toolHandler := NewDefaultToolHandler()
	server := &MCPServer{
		ToolHandler:  toolHandler,
		openAIClient: openai.NewClient(os.Getenv("OPENAI_API_KEY")),
	}

	fmt.Println("Setting up vector search...")
	// Setup vector search
	if err := server.SetupVectorSearch(); err != nil {
		log.Fatalf("Failed to setup vector search: %v", err)
	}

	fmt.Println("Setting up repository indexer...")
	// Setup repository indexer
	if err := server.SetupRepoIndexer(); err != nil {
		log.Fatalf("Failed to setup repository indexer: %v", err)
	}

	// Setup routes with all endpoints
	http.HandleFunc("/mcp-info", server.HandleMCPRegistration)
	http.HandleFunc("/health", server.HandleHealthCheck)
	http.HandleFunc("/vector-search", server.HandleVectorSearch)
	http.HandleFunc("/index-repository", server.HandleRepositoryIndexing)
	http.HandleFunc("/chat", server.HandleChat)
	http.HandleFunc("/github-config", server.HandleGitHubConfig)
	http.HandleFunc("/cursor", server.HandleCursorConnection)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Printf("MCP Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
