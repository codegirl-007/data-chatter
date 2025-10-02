// Package tools provides database query tools for LLM integration.
package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"data-chatter/internal/database"
	"data-chatter/internal/types"
)

// DatabaseQueryTool executes read-only SQL SELECT queries with security validation.
type DatabaseQueryTool struct {
	conn *database.Connection
}

// NewDatabaseQueryTool creates a new database query tool instance.
func NewDatabaseQueryTool(conn *database.Connection) *DatabaseQueryTool {
	return &DatabaseQueryTool{
		conn: conn,
	}
}

// GetDefinition returns the tool definition for LLM integration.
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

// Validate performs security checks on the SQL query to ensure only SELECT statements are allowed.
func (d *DatabaseQueryTool) Validate(input map[string]interface{}) error {
	query, ok := input["query"].(string)
	if !ok {
		return fmt.Errorf("query must be a string")
	}
	if query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(queryUpper, "SELECT") {
		return fmt.Errorf("only SELECT queries are allowed")
	}

	dangerousKeywords := []string{"DROP", "DELETE", "UPDATE", "INSERT", "ALTER", "CREATE", "TRUNCATE"}
	for _, keyword := range dangerousKeywords {
		if strings.Contains(queryUpper, keyword) {
			return fmt.Errorf("query contains forbidden keyword: %s", keyword)
		}
	}

	return nil
}

// Execute runs the SQL query and returns formatted results as JSON.
// Handles type conversion for different database column types.
func (d *DatabaseQueryTool) Execute(input map[string]interface{}) (*types.ToolResult, error) {
	query := input["query"].(string)

	fmt.Printf("DEBUG: Executing query: %s\n", query)

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

	var results []map[string]interface{}
	rowCount := 0

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
					Text: fmt.Sprintf("Failed to scan row: %v", err),
				}},
				IsError: true,
				Error:   &types.ToolError{Type: "query_error", Message: err.Error()},
			}, nil
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
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

	response := map[string]interface{}{
		"query":     query,
		"columns":   columns,
		"row_count": rowCount,
		"data":      results,
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
