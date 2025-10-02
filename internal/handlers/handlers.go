package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"data-chatter/internal/database"
	"data-chatter/internal/engine"
	"data-chatter/internal/types"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

var startTime = time.Now()

// Global tool engine instance (will be initialized in main)
var toolEngine *engine.ToolEngine

// InitializeToolEngine initializes the global tool engine
func InitializeToolEngine(dbConn *database.Connection) {
	toolEngine = engine.NewToolEngine(dbConn)
}

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uptime := time.Since(startTime)
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HomeHandler handles the root endpoint
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := APIResponse{
		Message: "Welcome to Data Chatter API",
		Data: map[string]interface{}{
			"version": "1.0.0",
			"endpoints": map[string]string{
				"health": "/health",
				"api":    "/api/",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// APIHandler handles API requests
func APIHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Message: "API endpoint reached",
		Data: map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ToolsHandler handles requests to list available tools
func ToolsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := toolEngine.GetAvailableTools()
	response := APIResponse{
		Message: "Available tools",
		Data:    tools,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ToolCallHandler handles tool execution requests
func ToolCallHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request types.ToolExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response := APIResponse{
			Message: "Invalid request format",
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if len(request.Tools) == 0 {
		response := APIResponse{
			Message: "No tools provided",
			Error:   "At least one tool must be provided",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Execute tools
	results := toolEngine.ExecuteTools(request.Tools)
	response := types.ToolExecutionResponse{
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SingleToolHandler handles single tool execution
func SingleToolHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var toolCall types.ToolCall
	if err := json.NewDecoder(r.Body).Decode(&toolCall); err != nil {
		response := APIResponse{
			Message: "Invalid request format",
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if toolCall.Name == "" {
		response := APIResponse{
			Message: "Tool name is required",
			Error:   "Tool name cannot be empty",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Execute single tool
	result, err := toolEngine.ExecuteTool(toolCall.Name, toolCall.Input)
	if err != nil {
		response := APIResponse{
			Message: "Tool execution failed",
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
