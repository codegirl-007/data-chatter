package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// TestRequest represents a test request
type TestRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// TestResponse represents the API response
type TestResponse struct {
	Query    string                   `json:"query"`
	Columns  []string                 `json:"columns"`
	RowCount int                      `json:"row_count"`
	Data     []map[string]interface{} `json:"data"`
}

func main() {
	baseURL := "http://localhost:8081"

	fmt.Println("🧪 Testing Data Chatter API")
	fmt.Println("==========================")

	// Test 1: Health check
	fmt.Println("\n1️⃣ Testing health check...")
	testHealth(baseURL)

	// Test 2: List available tools
	fmt.Println("\n2️⃣ Testing tools endpoint...")
	testTools(baseURL)

	// Test 3: Direct database query
	fmt.Println("\n3️⃣ Testing direct database query...")
	testDirectQuery(baseURL)

	// Test 4: Schema query
	fmt.Println("\n4️⃣ Testing schema query...")
	testSchema(baseURL)

	// Test 5: LLM tool execution
	fmt.Println("\n5️⃣ Testing LLM tool execution...")
	testLLMTool(baseURL)

	fmt.Println("\n✅ All tests completed!")
}

func testHealth(baseURL string) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Health check: %s\n", string(body))
}

func testTools(baseURL string) {
	resp, err := http.Get(baseURL + "/tools")
	if err != nil {
		fmt.Printf("❌ Tools request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Available tools: %s\n", string(body))
}

func testDirectQuery(baseURL string) {
	// Test query: Get contacts with specific criteria
	query := TestRequest{
		Query: "SELECT name, phone_number, days_available FROM contacts WHERE days_available LIKE '%Monday%' LIMIT 3",
		Limit: 3,
	}

	jsonData, _ := json.Marshal(query)
	resp, err := http.Post(baseURL+"/db/query", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Direct query failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result TestResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		fmt.Printf("Raw response: %s\n", string(body))
		return
	}

	fmt.Printf("✅ Query executed successfully!\n")
	fmt.Printf("📊 Found %d contacts available on Monday:\n", result.RowCount)

	for i, contact := range result.Data {
		if i < 3 { // Show first 3 results
			fmt.Printf("   %d. %s - %s (%s)\n",
				i+1,
				contact["name"],
				contact["phone_number"],
				contact["days_available"])
		}
	}
}

func testSchema(baseURL string) {
	query := map[string]string{"table_name": "contacts"}
	jsonData, _ := json.Marshal(query)

	resp, err := http.Post(baseURL+"/db/schema", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Schema query failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Schema information: %s\n", string(body))
}

func testLLMTool(baseURL string) {
	// Test LLM tool execution
	toolCall := map[string]interface{}{
		"id":   "test-1",
		"type": "tool_use",
		"name": "database_query",
		"input": map[string]interface{}{
			"query": "SELECT COUNT(*) as total_contacts FROM contacts",
		},
	}

	jsonData, _ := json.Marshal(toolCall)
	resp, err := http.Post(baseURL+"/tools/single", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ LLM tool execution failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ LLM tool result: %s\n", string(body))
}

// Helper function to start server in background
func startServer() {
	fmt.Println("🚀 Starting server in background...")
	// This would start the server, but for testing we assume it's already running
}
