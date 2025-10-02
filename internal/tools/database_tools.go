package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"data-chatter/internal/database"
	"data-chatter/internal/types"
)

// DatabaseQueryTool handles SQL SELECT queries
type DatabaseQueryTool struct {
	conn *database.Connection
}

// NewDatabaseQueryTool creates a new database query tool
func NewDatabaseQueryTool(conn *database.Connection) *DatabaseQueryTool {
	return &DatabaseQueryTool{
		conn: conn,
	}
}

func (d *DatabaseQueryTool) GetDefinition() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "database_query",
		Description: "Execute a read-only SQL SELECT query on the database",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "SQL SELECT query to execute (include LIMIT clause if needed)",
				},
			},
			"required": []string{"query"},
		},
	}
}

func (d *DatabaseQueryTool) Validate(input map[string]interface{}) error {
	query, ok := input["query"].(string)
	if !ok {
		return fmt.Errorf("query must be a string")
	}
	if query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	// Security check - only allow SELECT statements
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(queryUpper, "SELECT") {
		return fmt.Errorf("only SELECT queries are allowed")
	}

	// Check for dangerous keywords
	dangerousKeywords := []string{"DROP", "DELETE", "UPDATE", "INSERT", "ALTER", "CREATE", "TRUNCATE"}
	for _, keyword := range dangerousKeywords {
		if strings.Contains(queryUpper, keyword) {
			return fmt.Errorf("query contains forbidden keyword: %s", keyword)
		}
	}

	return nil
}

func (d *DatabaseQueryTool) Execute(input map[string]interface{}) (*types.ToolResult, error) {
	query := input["query"].(string)

	// Let the LLM have full control over the query - no automatic LIMIT addition
	fmt.Printf("DEBUG: Executing query: %s\n", query)

	// Execute query
	rows, err := d.conn.DB.Query(query)
	if err != nil {
		return &types.ToolResult{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Query execution failed: %v", err),
			}},
			IsError: true,
			Error:   &types.ToolError{Type: "query_error", Message: err.Error()},
		}, nil
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return &types.ToolResult{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to get column names: %v", err),
			}},
			IsError: true,
			Error:   &types.ToolError{Type: "query_error", Message: err.Error()},
		}, nil
	}

	// Process rows
	var results []map[string]interface{}
	rowCount := 0

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			return &types.ToolResult{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Failed to scan row: %v", err),
				}},
				IsError: true,
				Error:   &types.ToolError{Type: "query_error", Message: err.Error()},
			}, nil
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				// Handle different data types
				switch v := val.(type) {
				case []byte:
					row[col] = string(v)
				case time.Time:
					row[col] = v.Format(time.RFC3339)
				default:
					row[col] = v
				}
			} else {
				row[col] = nil
			}
		}
		results = append(results, row)
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return &types.ToolResult{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error iterating rows: %v", err),
			}},
			IsError: true,
			Error:   &types.ToolError{Type: "query_error", Message: err.Error()},
		}, nil
	}

	// Create response
	response := map[string]interface{}{
		"query":     query,
		"columns":   columns,
		"row_count": rowCount,
		"data":      results,
	}

	// Debug: Print the query results
	fmt.Printf("DEBUG: Query results:\n")
	fmt.Printf("  Query: %s\n", query)
	fmt.Printf("  Columns: %v\n", columns)
	fmt.Printf("  Row count: %d\n", rowCount)
	fmt.Printf("  Results: %v\n\n", results)

	jsonData, _ := json.MarshalIndent(response, "", "  ")

	return &types.ToolResult{
		Content: []types.ToolContent{{
			Type: "text",
			Text: string(jsonData),
		}},
		IsError: false,
	}, nil
}

// DatabaseSchemaTool handles schema queries
type DatabaseSchemaTool struct {
	conn *database.Connection
}

// NewDatabaseSchemaTool creates a new database schema tool
func NewDatabaseSchemaTool(conn *database.Connection) *DatabaseSchemaTool {
	return &DatabaseSchemaTool{
		conn: conn,
	}
}

func (d *DatabaseSchemaTool) GetDefinition() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "database_schema",
		Description: "Get database schema information (tables, columns, etc.)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"table_name": map[string]interface{}{
					"type":        "string",
					"description": "Specific table name to get schema for (optional)",
				},
			},
		},
	}
}

func (d *DatabaseSchemaTool) Validate(input map[string]interface{}) error {
	// No validation needed for schema queries
	return nil
}

func (d *DatabaseSchemaTool) Execute(input map[string]interface{}) (*types.ToolResult, error) {
	tableName, hasTable := input["table_name"].(string)

	var query string
	var args []interface{}

	if hasTable && tableName != "" {
		// Get schema for specific table (SQLite syntax)
		query = `PRAGMA table_info(` + tableName + `)`
	} else {
		// Get all tables (SQLite syntax)
		query = `SELECT name as table_name FROM sqlite_master WHERE type='table' ORDER BY name`
	}

	rows, err := d.conn.DB.Query(query, args...)
	if err != nil {
		return &types.ToolResult{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Schema query failed: %v", err),
			}},
			IsError: true,
			Error:   &types.ToolError{Type: "query_error", Message: err.Error()},
		}, nil
	}
	defer rows.Close()

	var results []map[string]interface{}
	columns, _ := rows.Columns()

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return &types.ToolResult{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Failed to scan schema row: %v", err),
				}},
				IsError: true,
				Error:   &types.ToolError{Type: "query_error", Message: err.Error()},
			}, nil
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				if v, ok := val.([]byte); ok {
					row[col] = string(v)
				} else {
					row[col] = val
				}
			} else {
				row[col] = nil
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return &types.ToolResult{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error iterating schema rows: %v", err),
			}},
			IsError: true,
			Error:   &types.ToolError{Type: "query_error", Message: err.Error()},
		}, nil
	}

	// Create a proper response structure
	response := map[string]interface{}{
		"tables": results,
		"count":  len(results),
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")

	return &types.ToolResult{
		Content: []types.ToolContent{{
			Type: "text",
			Text: string(jsonData),
		}},
		IsError: false,
	}, nil
}
