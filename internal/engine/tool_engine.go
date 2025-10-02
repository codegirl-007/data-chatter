package engine

import (
	"data-chatter/internal/database"
	"data-chatter/internal/tools"
	"data-chatter/internal/types"
)

// ToolEngine manages tool execution and provides a centralized interface
type ToolEngine struct {
	registry *types.ToolRegistry
}

// NewToolEngine creates a new tool engine with all tools registered
func NewToolEngine(dbConn *database.Connection) *ToolEngine {
	engine := &ToolEngine{
		registry: types.NewToolRegistry(),
	}

	// Register all available tools
	engine.registerTools(dbConn)

	return engine
}

// registerTools registers all available tools with the registry
func (te *ToolEngine) registerTools(dbConn *database.Connection) {
	// Database tools
	te.registry.RegisterTool("database_query", tools.NewDatabaseQueryTool(dbConn))
	te.registry.RegisterTool("database_schema", tools.NewDatabaseSchemaTool(dbConn))
	te.registry.RegisterTool("database_smart_query", tools.NewDatabaseSmartQueryTool(dbConn))
}

// ExecuteTools executes a list of tool calls
func (te *ToolEngine) ExecuteTools(toolCalls []types.ToolCall) []types.ToolResult {
	return te.registry.ExecuteTools(toolCalls)
}

// ExecuteTool executes a single tool
func (te *ToolEngine) ExecuteTool(name string, input map[string]interface{}) (*types.ToolResult, error) {
	return te.registry.ExecuteTool(name, input)
}

// GetAvailableTools returns all available tools
func (te *ToolEngine) GetAvailableTools() []types.ToolDefinition {
	return te.registry.ListTools()
}

// GetTool returns a specific tool by name
func (te *ToolEngine) GetTool(name string) (types.ToolRegistryEntry, bool) {
	return te.registry.GetTool(name)
}
