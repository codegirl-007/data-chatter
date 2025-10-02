package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"data-chatter/internal/database"
)

// AnthropicClient handles communication with Anthropic API
type AnthropicClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	DB         *database.Connection
}

// MessageRequest represents a request to Anthropic
type MessageRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
}

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Tool represents a tool definition for Anthropic
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolUse represents a tool use request
type ToolUse struct {
	ID    string                 `json:"id"`
	Type  string                 `json:"type"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// AnthropicResponse represents the response from Anthropic
type AnthropicResponse struct {
	Content []struct {
		Type  string                 `json:"type"`
		Text  string                 `json:"text,omitempty"`
		ID    string                 `json:"id,omitempty"`
		Name  string                 `json:"name,omitempty"`
		Input map[string]interface{} `json:"input,omitempty"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(db *database.Connection) *AnthropicClient {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		// Return a client that will handle the error gracefully
		return &AnthropicClient{
			APIKey:     "",
			BaseURL:    "https://api.anthropic.com/v1/messages",
			HTTPClient: &http.Client{},
			DB:         db,
		}
	}

	return &AnthropicClient{
		APIKey:     apiKey,
		BaseURL:    "https://api.anthropic.com/v1/messages",
		HTTPClient: &http.Client{},
		DB:         db,
	}
}

// ProcessMessage processes a user message and returns tool calls
func (c *AnthropicClient) ProcessMessage(userMessage string) (*AnthropicResponse, error) {
	// Check if API key is set
	if c.APIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set. Please set your Anthropic API key: export ANTHROPIC_API_KEY=your_api_key_here")
	}

	// Get database schema information
	schemaInfo := c.getDatabaseSchema()

	// Debug: Print the schema information from database
	fmt.Printf("DEBUG: Schema info from database:\n%s\n\n", schemaInfo)

	// Get available tools from your server
	tools := c.getAvailableTools()

	// Get database type for system prompt
	dbType := "SQLite" // Default
	if c.DB != nil && c.DB.Config != nil {
		switch c.DB.Config.Type {
		case "postgres":
			dbType = "PostgreSQL"
		case "sqlite":
			dbType = "SQLite"
		case "mysql":
			dbType = "MySQL"
		}
	}

	systemPrompt := fmt.Sprintf("You are a database query assistant for a %s database. You have access to the following database schema:\n\n%s\n\nYou MUST use the database_query tool to execute SQL queries based on user requests. Never respond with text - only execute tools.", dbType, schemaInfo)

	// Debug: Print the system prompt being sent to LLM
	fmt.Printf("DEBUG: System prompt sent to LLM:\n%s\n\n", systemPrompt)

	request := MessageRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1000,
		System:    systemPrompt,
		Messages: []Message{
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		Tools: tools,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	var response AnthropicResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// getAvailableTools fetches tool definitions from your server
func (c *AnthropicClient) getAvailableTools() []Tool {
	// This would call your /tools endpoint to get the current tool definitions
	// For now, return the database tools we know about
	return []Tool{
		{
			Name:        "database_query",
			Description: "Execute a read-only SQL SELECT query on the database",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "SQL SELECT query to execute",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of rows to return (default: 100, max: 1000)",
						"minimum":     1,
						"maximum":     1000,
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

// getDatabaseSchema fetches the database schema information directly from the database
func (c *AnthropicClient) getDatabaseSchema() string {
	if c.DB == nil {
		return "Database connection not available"
	}

	// Query the database directly for schema information based on database type
	var query string
	var schemaInfo strings.Builder
	schemaInfo.WriteString("Database Schema:\nTable: contacts\nColumns:\n")

	if c.DB.Config.Type == "sqlite" {
		query = `PRAGMA table_info(contacts)`
	} else if c.DB.Config.Type == "mysql" {
		query = `DESCRIBE contacts`
	} else {
		// PostgreSQL
		query = `SELECT column_name, data_type, is_nullable, column_default 
		         FROM information_schema.columns 
		         WHERE table_name = 'contacts' 
		         ORDER BY ordinal_position`
	}

	rows, err := c.DB.DB.Query(query)
	if err != nil {
		return "Failed to get database schema"
	}
	defer rows.Close()

	if c.DB.Config.Type == "sqlite" {
		// SQLite schema parsing
		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull, pk int
			var dfltValue interface{}

			err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk)
			if err != nil {
				continue
			}

			nullable := "NULL"
			if notNull == 1 {
				nullable = "NOT NULL"
			}

			primaryKey := ""
			if pk == 1 {
				primaryKey = ", PRIMARY KEY"
			}

			schemaInfo.WriteString(fmt.Sprintf("- %s (%s, %s%s)\n", name, dataType, nullable, primaryKey))
		}
	} else if c.DB.Config.Type == "mysql" {
		// MySQL schema parsing
		for rows.Next() {
			var field, dataType, null, key, defaultValue, extra string

			err := rows.Scan(&field, &dataType, &null, &key, &defaultValue, &extra)
			if err != nil {
				continue
			}

			nullable := "NULL"
			if null == "NO" {
				nullable = "NOT NULL"
			}

			primaryKey := ""
			if key == "PRI" {
				primaryKey = ", PRIMARY KEY"
			}

			schemaInfo.WriteString(fmt.Sprintf("- %s (%s, %s%s)\n", field, dataType, nullable, primaryKey))
		}
	} else {
		// PostgreSQL schema parsing
		for rows.Next() {
			var columnName, dataType, isNullable, columnDefault string

			err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
			if err != nil {
				continue
			}

			nullable := "NULL"
			if isNullable == "NO" {
				nullable = "NOT NULL"
			}

			schemaInfo.WriteString(fmt.Sprintf("- %s (%s, %s)\n", columnName, dataType, nullable))
		}
	}

	schemaInfo.WriteString("\nThe days_available column contains comma-separated values like \"Monday, Tuesday, Wednesday\".")

	return schemaInfo.String()
}
