package handlers

import (
	"encoding/json"
	"net/http"

	"data-chatter/internal/database"
	"data-chatter/internal/tools"
)

// DatabaseHandler handles direct database queries
type DatabaseHandler struct {
	queryTool  *tools.DatabaseQueryTool
	schemaTool *tools.DatabaseSchemaTool
}

// NewDatabaseHandler creates a new database handler
func NewDatabaseHandler(conn *database.Connection) *DatabaseHandler {
	return &DatabaseHandler{
		queryTool:  tools.NewDatabaseQueryTool(conn),
		schemaTool: tools.NewDatabaseSchemaTool(conn),
	}
}

// QueryRequest represents a database query request
type QueryRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// SchemaRequest represents a schema query request
type SchemaRequest struct {
	TableName string `json:"table_name,omitempty"`
}

// QueryHandler handles direct database queries
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

	// Set default limit if not provided
	if request.Limit == 0 {
		request.Limit = 100
	}

	// Execute the query
	input := map[string]interface{}{
		"query": request.Query,
		"limit": request.Limit,
	}

	result, err := dh.queryTool.Execute(input)
	if err != nil {
		http.Error(w, "Query execution failed", http.StatusInternalServerError)
		return
	}

	// Return the raw data directly (not wrapped in tool result)
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

// SchemaHandler handles schema queries
func (dh *DatabaseHandler) SchemaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request SchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Execute the schema query
	input := map[string]interface{}{}
	if request.TableName != "" {
		input["table_name"] = request.TableName
	}

	result, err := dh.schemaTool.Execute(input)
	if err != nil {
		http.Error(w, "Schema query failed", http.StatusInternalServerError)
		return
	}

	// Return the raw data directly
	if len(result.Content) > 0 {
		// For schema queries, return the raw text result
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result.Content[0].Text))
	} else {
		http.Error(w, "No schema data returned", http.StatusInternalServerError)
	}
}
