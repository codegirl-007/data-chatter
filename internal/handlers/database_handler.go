// Package handlers provides HTTP request handlers for direct database access.
package handlers

import (
	"encoding/json"
	"net/http"

	"data-chatter/internal/database"
	"data-chatter/internal/tools"
)

// DatabaseHandler provides direct database query access for API clients.
type DatabaseHandler struct {
	queryTool *tools.DatabaseQueryTool
}

// NewDatabaseHandler creates a new database handler with query tool.
func NewDatabaseHandler(conn *database.Connection) *DatabaseHandler {
	return &DatabaseHandler{
		queryTool: tools.NewDatabaseQueryTool(conn),
	}
}

// QueryRequest represents a database query request.
type QueryRequest struct {
	Query string `json:"query"`
}

// QueryHandler executes direct database queries and returns results as JSON.
func (dh *DatabaseHandler) QueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if request.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	input := map[string]interface{}{
		"query": request.Query,
	}

	result, err := dh.queryTool.Execute(input)
	if err != nil {
		http.Error(w, "Query execution failed", http.StatusInternalServerError)
		return
	}

	if len(result.Content) > 0 {
		var data interface{}
		if err := json.Unmarshal([]byte(result.Content[0].Text), &data); err != nil {
			http.Error(w, "Failed to parse query result", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	} else {
		http.Error(w, "No data returned", http.StatusInternalServerError)
	}
}

// SchemaHandler returns a simple message since schema is now handled by LLM client.
func (dh *DatabaseHandler) SchemaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]string{
		"message": "Schema information is now provided directly to the LLM client",
		"note":    "Use /llm/message endpoint for LLM integration with schema",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
