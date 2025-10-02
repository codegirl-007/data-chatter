package types

import (
	"fmt"
)

// ToolCall represents a tool call request from Claude
type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Input    map[string]interface{} `json:"input"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ID      string        `json:"id"`
	Content []ToolContent `json:"content"`
	IsError bool          `json:"is_error"`
	Error   *ToolError    `json:"error,omitempty"`
	Usage   *ToolUsage    `json:"usage,omitempty"`
}

// ToolContent represents content in a tool result
type ToolContent struct {
	Type string      `json:"type"`
	Text string      `json:"text,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// ToolError represents an error in tool execution
type ToolError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ToolUsage represents usage statistics for a tool
type ToolUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ToolDefinition represents the definition of a tool
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolExecutionRequest represents a request to execute tools
type ToolExecutionRequest struct {
	Tools []ToolCall `json:"tools"`
}

// ToolExecutionResponse represents the response from tool execution
type ToolExecutionResponse struct {
	Results []ToolResult `json:"results"`
}

// ToolRegistryEntry represents an entry in the tool registry
type ToolRegistryEntry struct {
	Definition ToolDefinition
	Executor   ToolExecutor
}

// ToolExecutor is the interface that all tools must implement
type ToolExecutor interface {
	Execute(input map[string]interface{}) (*ToolResult, error)
	GetDefinition() ToolDefinition
	Validate(input map[string]interface{}) error
}

// ToolRegistry manages available tools
type ToolRegistry struct {
	tools map[string]ToolRegistryEntry
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ToolRegistryEntry),
	}
}

// RegisterTool registers a new tool
func (tr *ToolRegistry) RegisterTool(name string, executor ToolExecutor) {
	tr.tools[name] = ToolRegistryEntry{
		Definition: executor.GetDefinition(),
		Executor:   executor,
	}
}

// GetTool retrieves a tool by name
func (tr *ToolRegistry) GetTool(name string) (ToolRegistryEntry, bool) {
	tool, exists := tr.tools[name]
	return tool, exists
}

// ListTools returns all registered tools
func (tr *ToolRegistry) ListTools() []ToolDefinition {
	definitions := make([]ToolDefinition, 0, len(tr.tools))
	for _, entry := range tr.tools {
		definitions = append(definitions, entry.Definition)
	}
	return definitions
}

// ExecuteTool executes a tool by name
func (tr *ToolRegistry) ExecuteTool(name string, input map[string]interface{}) (*ToolResult, error) {
	entry, exists := tr.GetTool(name)
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	// Validate input
	if err := entry.Executor.Validate(input); err != nil {
		return &ToolResult{
			ID:      input["id"].(string),
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Validation error: %v", err)}},
			IsError: true,
			Error:   &ToolError{Type: "validation_error", Message: err.Error()},
		}, nil
	}

	// Execute tool
	return entry.Executor.Execute(input)
}

// ExecuteTools executes multiple tools
func (tr *ToolRegistry) ExecuteTools(toolCalls []ToolCall) []ToolResult {
	results := make([]ToolResult, len(toolCalls))

	for i, toolCall := range toolCalls {
		result, err := tr.ExecuteTool(toolCall.Name, toolCall.Input)
		if err != nil {
			results[i] = ToolResult{
				ID:      toolCall.ID,
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Execution error: %v", err)}},
				IsError: true,
				Error:   &ToolError{Type: "execution_error", Message: err.Error()},
			}
		} else {
			results[i] = *result
		}
	}

	return results
}
