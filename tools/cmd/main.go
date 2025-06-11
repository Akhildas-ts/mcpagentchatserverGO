package main

import (
	"log"
	"net/http"
	"os"

	"mcpserver/internal/config"
	"mcpserver/internal/handlers"
	"mcpserver/internal/services"
	"mcpserver/internal/storage"

	"github.com/joho/godotenv"
)

type Server struct {
	Config   *config.Config
	Services *Services
	Handlers *Handlers
}

type Services struct {
	VectorSearch *services.VectorSearchService
	RepoIndexer  *services.RepoIndexerService
	MCPServer    *services.MCPServerService
}

type Handlers struct {
	Health       *handlers.HealthHandler
	VectorSearch *handlers.VectorSearchHandler
	RepoIndexer  *handlers.RepoIndexerHandler
	MCP          *handlers.MCPHandler
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize server
	server, err := initializeServer()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	// Setup routes
	router := setupRoutes(server.Handlers)

	// Start server
	log.Printf("MCP Server starting on port %s...", server.Config.Port)
	if err := http.ListenAndServe(":"+server.Config.Port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func initializeServer() (*Server, error) {
	// Load configuration
	cfg := config.Load()

	// Initialize storage layer
	pineconeStore, err := storage.NewPineconeStore(
		cfg.PineconeAPIKey,
		cfg.PineconeEnvironment,
		cfg.PineconeIndexName,
		cfg.PineconeHost,
	)
	if err != nil {
		return nil, err
	}

	openaiClient := storage.NewOpenAIClient(cfg.OpenAIAPIKey)

	// Initialize services
	services := &Services{
		VectorSearch: services.NewVectorSearchService(pineconeStore, openaiClient),
		RepoIndexer:  services.NewRepoIndexerService(pineconeStore, openaiClient),
		MCPServer:    services.NewMCPServerService(pineconeStore, openaiClient),
	}

	// Initialize handlers
	handlers := &Handlers{
		Health:       handlers.NewHealthHandler(),
		VectorSearch: handlers.NewVectorSearchHandler(services.VectorSearch),
		RepoIndexer:  handlers.NewRepoIndexerHandler(services.RepoIndexer),
		MCP:          handlers.NewMCPHandler(services.MCPServer),
	}

	return &Server{
		Config:   cfg,
		Services: services,
		Handlers: handlers,
	}, nil
}

func setupRoutes(h *Handlers) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", h.Health.HandleHealthCheck)

	// MCP endpoints
	mux.HandleFunc("/mcp-info", h.MCP.HandleMCPRegistration)
	mux.HandleFunc("/cursor", h.MCP.HandleCursorConnection)
	mux.HandleFunc("/chat", h.MCP.HandleChat)
	mux.HandleFunc("/github-config", h.MCP.HandleGitHubConfig)

	// Vector search endpoints
	mux.HandleFunc("/vector-search", h.VectorSearch.HandleVectorSearch)

	// Repository indexing endpoints
	mux.HandleFunc("/index-repository", h.RepoIndexer.HandleRepositoryIndexing)

	return mux
}