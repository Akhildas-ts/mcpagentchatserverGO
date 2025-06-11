package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"mcpserver/internal/models"
	"mcpserver/internal/service"
)

type MCPHandler struct {
	service *service.MCPServerService
}

func NewMCPHandler(service *service.MCPServerService) *MCPHandler {
	return &MCPHandler{
		service: service,
	}
}

func (h *MCPHandler) HandleMCPRegistration(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	serverInfo := h.service.GetServerInfo()
	sendMCPResponse(w, true, serverInfo, "")
}

func (h *MCPHandler) HandleCursorConnection(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req models.CursorRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendMCPResponse(w, false, nil, "Invalid cursor request format")
		return
	}

	// Handle cursor action
	result, err := h.service.HandleCursorAction(req.Action, req.Data)
	if err != nil {
		sendMCPResponse(w, false, nil, fmt.Sprintf("Cursor action failed: %v", err))
		return
	}

	sendMCPResponse(w, true, result, "")
}

func (h *MCPHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req models.ChatRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendMCPResponse(w, false, nil, "Invalid chat request format")
		return
	}

	if req.Message == "" || req.Repository == "" {
		sendMCPResponse(w, false, nil, "Message and repository are required")
		return
	}

	// Handle chat
	result, err := h.service.HandleChat(req.Message, req.Repository, req.Context)
	if err != nil {
		sendMCPResponse(w, false, nil, fmt.Sprintf("Chat failed: %v", err))
		return
	}

	sendMCPResponse(w, true, result, "")
}

func (h *MCPHandler) HandleGitHubConfig(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req models.GitHubConfigRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendMCPResponse(w, false, nil, "Invalid configuration request")
		return
	}

	if req.Repository == "" || req.Token == "" {
		sendMCPResponse(w, false, nil, "Repository and token are required")
		return
	}

	// Configure GitHub
	err := h.service.ConfigureGitHub(req.Repository, req.Token)
	if err != nil {
		sendMCPResponse(w, false, nil, fmt.Sprintf("GitHub configuration failed: %v", err))
		return
	}

	result := map[string]string{"status": "configured"}
	sendMCPResponse(w, true, result, "GitHub repository configured")
}

func sendMCPResponse(w http.ResponseWriter, success bool, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	response := &models.APIResponse{
		Success: success,
		Data:    data,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}