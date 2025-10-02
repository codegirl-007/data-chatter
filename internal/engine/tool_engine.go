// Package engine provides tool execution management for LLM integration.
package engine

import (
	"data-chatter/internal/database"
	"data-chatter/internal/tools"
	"data-chatter/internal/types"
)

// ToolEngine manages tool registration and execution for LLM tool calls.
type ToolEngine struct {
	registry *types.ToolRegistry
}

// NewToolEngine creates a new tool engine and registers all available tools.
func NewToolEngine(dbConn *database.Connection) *ToolEngine {
	engine := &ToolEngine{
		registry: types.NewToolRegistry(),
	}

	engine.registerTools(dbConn)

	return engine
}

// registerTools registers all available tools with the tool registry.
func (te *ToolEngine) registerTools(dbConn *database.Connection) {
	te.registry.RegisterTool("database_query", tools.NewDatabaseQueryTool(dbConn))
}

// ExecuteTools executes multiple tool calls and returns their results.
func (te *ToolEngine) ExecuteTools(toolCalls []types.ToolCall) []types.ToolResult {
	return te.registry.ExecuteTools(toolCalls)
}

// ExecuteTool executes a single tool by name with the provided input parameters.
func (te *ToolEngine) ExecuteTool(name string, input map[string]interface{}) (*types.ToolResult, error) {
	return te.registry.ExecuteTool(name, input)
}

// GetAvailableTools returns definitions for all registered tools.
func (te *ToolEngine) GetAvailableTools() []types.ToolDefinition {
	return te.registry.ListTools()
}
