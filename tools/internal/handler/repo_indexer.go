package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"mcpserver/internal/models"
	"mcpserver/internal/service"
)

type RepoIndexerHandler struct {
	service *service.RepoIndexerService
}

func NewRepoIndexerHandler(service *service.RepoIndexerService) *RepoIndexerHandler {
	return &RepoIndexerHandler{
		service: service,
	}
}

func (h *RepoIndexerHandler) HandleRepositoryIndexing(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req models.IndexRepositoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponseError(w, "Invalid request format")
		return
	}

	if req.RepoURL == "" {
		sendResponseError(w, "Repository URL is required")
		return
	}

	// Set default branch if not provided
	if req.Branch == "" {
		req.Branch = "main"
	}

	// Index repository
	err := h.service.IndexRepository(req.RepoURL, req.Branch)
	if err != nil {
		sendResponseError(w, fmt.Sprintf("Repository indexing failed: %v", err))
		return
	}

	result := map[string]interface{}{
		"status":  "success",
		"message": "Repository indexed successfully",
	}

	sendResponseSuccess(w, result, "Repository indexed successfully")
}

func sendResponseSuccess(w http.ResponseWriter, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	response := &models.APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}

func sendResponseError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	response := &models.APIResponse{
		Success: false,
		Data:    nil,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}