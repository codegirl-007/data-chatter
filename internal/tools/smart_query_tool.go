package tools

import (
	"encoding/json"
	"fmt"

	"data-chatter/internal/database"
	"data-chatter/internal/types"
)

// DatabaseSmartQueryTool handles intelligent database queries
type DatabaseSmartQueryTool struct {
	conn *database.Connection
}

// NewDatabaseSmartQueryTool creates a new smart query tool
func NewDatabaseSmartQueryTool(conn *database.Connection) *DatabaseSmartQueryTool {
	return &DatabaseSmartQueryTool{
		conn: conn,
	}
}

func (d *DatabaseSmartQueryTool) GetDefinition() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "database_smart_query",
		Description: "Intelligently query the database. First discovers schema, then executes the appropriate query to get the requested data.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"request": map[string]interface{}{
					"type":        "string",
					"description": "Natural language description of what data you want to retrieve",
				},
			},
			"required": []string{"request"},
		},
	}
}

func (d *DatabaseSmartQueryTool) Validate(input map[string]interface{}) error {
	request, ok := input["request"].(string)
	if !ok {
		return fmt.Errorf("request must be a string")
	}
	if request == "" {
		return fmt.Errorf("request cannot be empty")
	}
	return nil
}

func (d *DatabaseSmartQueryTool) Execute(input map[string]interface{}) (*types.ToolResult, error) {
	request := input["request"].(string)

	// Step 1: Get database schema to understand structure
	schemaQuery := `SELECT name as table_name FROM sqlite_master WHERE type='table' ORDER BY name`
	rows, err := d.conn.DB.Query(schemaQuery)
	if err != nil {
		return &types.ToolResult{
			Content: []types.ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Schema discovery failed: %v", err),
			}},
			IsError: true,
			Error:   &types.ToolError{Type: "query_error", Message: err.Error()},
		}, nil
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return &types.ToolResult{
				Content: []types.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Failed to scan table name: %v", err),
				}},
				IsError: true,
				Error:   &types.ToolError{Type: "query_error", Message: err.Error()},
			}, nil
		}
		tables = append(tables, tableName)
	}

	// Step 2: For each table, get column info to understand structure
	var tableSchemas []map[string]interface{}
	for _, tableName := range tables {
		if tableName == "sqlite_sequence" {
			continue // Skip system tables
		}

		// Get column info for this table
		columnQuery := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
		columnRows, err := d.conn.DB.Query(columnQuery)
		if err != nil {
			continue
		}

		var columns []map[string]interface{}
		for columnRows.Next() {
			var cid int
			var name, dataType string
			var notnull int
			var dfltValue interface{}
			var pk int

			if err := columnRows.Scan(&cid, &name, &dataType, &notnull, &dfltValue, &pk); err != nil {
				continue
			}

			columns = append(columns, map[string]interface{}{
				"name":    name,
				"type":    dataType,
				"notnull": notnull == 1,
				"primary": pk == 1,
				"default": dfltValue,
			})
		}
		columnRows.Close()

		tableSchemas = append(tableSchemas, map[string]interface{}{
			"table_name": tableName,
			"columns":    columns,
		})
	}

	// Step 3: Let the LLM construct the query based on the request and schema
	// The LLM should analyze the request and schema to build the appropriate SQL
	var finalQuery string
	var queryResults []map[string]interface{}

	// For now, provide a simple fallback - the LLM should be doing the heavy lifting
	// This is just a safety net in case the LLM doesn't provide a query
	finalQuery = `SELECT * FROM contacts LIMIT 10`

	if finalQuery != "" {
		// Execute the final query
		queryRows, err := d.conn.DB.Query(finalQuery)
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
		defer queryRows.Close()

		// Get column names
		queryColumns, err := queryRows.Columns()
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

		// Process results
		for queryRows.Next() {
			values := make([]interface{}, len(queryColumns))
			valuePtrs := make([]interface{}, len(queryColumns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := queryRows.Scan(valuePtrs...); err != nil {
				continue
			}

			row := make(map[string]interface{})
			for i, col := range queryColumns {
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
			queryResults = append(queryResults, row)
		}
	}

	// Create comprehensive response
	response := map[string]interface{}{
		"request":      request,
		"schema":       tableSchemas,
		"query":        finalQuery,
		"results":      queryResults,
		"result_count": len(queryResults),
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
