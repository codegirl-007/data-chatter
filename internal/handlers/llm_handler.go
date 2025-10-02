package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"data-chatter/internal/database"
	"data-chatter/internal/llm"
)

// LLMHandler handles LLM integration requests
type LLMHandler struct {
	anthropicClient *llm.AnthropicClient
}

// NewLLMHandler creates a new LLM handler
func NewLLMHandler(db *database.Connection) *LLMHandler {
	return &LLMHandler{
		anthropicClient: llm.NewAnthropicClient(db),
	}
}

// MessageRequest represents a message from the UI
type MessageRequest struct {
	Message string `json:"message"`
}

// MessageResponse represents the response to the UI
type MessageResponse struct {
	Message string      `json:"message"`
	Results interface{} `json:"results,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ProcessMessageHandler handles message processing with LLM
func (lh *LLMHandler) ProcessMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response := MessageResponse{
			Message: "Invalid request format",
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if request.Message == "" {
		response := MessageResponse{
			Message: "Message is required",
			Error:   "Message cannot be empty",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Process message with Anthropic
	anthropicResponse, err := lh.anthropicClient.ProcessMessage(request.Message)
	if err != nil {
		// Check if it's an API key error
		if strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
			response := MessageResponse{
				Message: "❌ Anthropic API key not configured",
				Error:   err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := MessageResponse{
			Message: "Failed to process message with LLM",
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if LLM wants to use tools
	if len(anthropicResponse.Content) > 0 && anthropicResponse.Content[0].Type == "tool_use" {
		// Debug: Log how many tool calls we received
		fmt.Printf("DEBUG: Received %d tool calls from LLM\n", len(anthropicResponse.Content))

		// Execute all tool calls in sequence
		var allResults []interface{}
		var lastError error

		for i, content := range anthropicResponse.Content {
			if content.Type == "tool_use" {
				fmt.Printf("DEBUG: Executing tool call %d: %s\n", i+1, content.Name)
				results, err := lh.executeToolCall(content)
				if err != nil {
					lastError = err
					break
				}
				allResults = append(allResults, results)
			}
		}

		if lastError != nil {
			response := MessageResponse{
				Message: "Failed to execute tool call",
				Error:   lastError.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Return results directly to UI
		response := MessageResponse{
			Message: "Query executed successfully",
			Results: allResults,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// If no tool use, return the text response
	response := MessageResponse{
		Message: anthropicResponse.Content[0].Text,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// executeToolCall executes a tool call and returns the results
func (lh *LLMHandler) executeToolCall(toolUseContent struct {
	Type  string                 `json:"type"`
	Text  string                 `json:"text,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}) (interface{}, error) {
	// Convert Anthropic tool use to our tool call format
	toolCall := map[string]interface{}{
		"id":    toolUseContent.ID,
		"type":  "tool_use",
		"name":  toolUseContent.Name,
		"input": toolUseContent.Input,
	}

	// Execute the tool call using our existing tool system
	jsonData, _ := json.Marshal(toolCall)

	// Make HTTP call to our own tool execution endpoint
	resp, err := http.Post("http://localhost:8081/tools/single", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool call: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tool result: %w", err)
	}

	return result, nil
}
