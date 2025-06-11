package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"mcpserver/internal/models"
	"mcpserver/internal/service"
)

type VectorSearchHandler struct {
	service *service.VectorSearchService
}

func NewVectorSearchHandler(service *service.VectorSearchService) *VectorSearchHandler {
	return &VectorSearchHandler{
		service: service,
	}
}

func (h *VectorSearchHandler) HandleVectorSearch(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
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

	// Create search request
	searchRequest := &models.SearchRequest{
		Query:      req.Query,
		Repository: req.Repository,
		Branch:     req.Branch,
		Limit:      10,
	}

	// Execute vector search with summary
	result, err := h.service.SearchWithSummary(searchRequest)
	if err != nil {
		sendResponse(w, false, nil, fmt.Sprintf("Search failed: %v", err))
		return
	}

	sendResponse(w, true, result, "")
}

func sendResponse(w http.ResponseWriter, success bool, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	response := &models.APIResponse{
		Success: success,
		Data:    data,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}