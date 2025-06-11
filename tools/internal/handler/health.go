package handler

import (
	"encoding/json"
	"net/http"

	"mcpserver/internal/models"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	response := &models.APIResponse{
		Success: true,
		Data:    map[string]string{"status": "healthy"},
		Message: "",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}