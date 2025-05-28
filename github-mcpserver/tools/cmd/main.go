// cmd/main.go in tools directory
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// ServerInfo represents the MCP server information
type ServerInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Vendor      struct {
		Name string `json:"name"`
	} `json:"vendor"`
	Capabilities struct {
		VectorSearch    bool `json:"vectorSearch"`
		IndexRepository bool `json:"indexRepository"`
		Chat            bool `json:"chat"`
	} `json:"capabilities"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	// Configure server with longer timeouts to prevent client closed errors
	server := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 60 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Create server info for MCP registration
	serverInfo := ServerInfo{
		Name:        "Go Vector Search Server",
		Version:     "1.0.0",
		Description: "Backend for vector search and repository indexing",
		Vendor: struct {
			Name string `json:"name"`
		}{
			Name: "Your Organization",
		},
	}
	serverInfo.Capabilities.VectorSearch = true
	serverInfo.Capabilities.IndexRepository = true
	serverInfo.Capabilities.Chat = true

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// MCP registration endpoint
	http.HandleFunc("/mcp-registration", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := struct {
			ServerInfo ServerInfo `json:"serverInfo"`
			Status     string     `json:"status"`
		}{
			ServerInfo: serverInfo,
			Status:     "success",
		}
		json.NewEncoder(w).Encode(response)
	})

	// Vector search endpoint (placeholder)
	http.HandleFunc("/vector-search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
			return
		}

		// TODO: Implement actual vector search
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"results": []map[string]string{
				{"filename": "example.go", "snippet": "func ExampleFunction() {...}"},
			},
		})
	})

	// Index repository endpoint (placeholder)
	http.HandleFunc("/index-repository", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
			return
		}

		// TODO: Implement actual repository indexing
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Repository indexing started",
		})
	})

	// Chat endpoint (placeholder)
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
			return
		}

		// TODO: Implement actual chat processing
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "success",
			"response": "This is a placeholder response from the Go server",
		})
	})

	// Start the server
	log.Printf("Starting server on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
