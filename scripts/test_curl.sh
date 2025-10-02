#!/bin/bash

echo "🧪 Testing Data Chatter API with curl"
echo "===================================="

BASE_URL="http://localhost:8081"

# Test 1: Health check
echo -e "\n1️⃣ Testing health check..."
curl -s "$BASE_URL/health" | jq '.' || echo "Health check response received"

# Test 2: List available tools
echo -e "\n2️⃣ Testing tools endpoint..."
curl -s "$BASE_URL/tools" | jq '.'

# Test 3: Direct database query - Get contacts available on Monday
echo -e "\n3️⃣ Testing direct database query..."
curl -s -X POST "$BASE_URL/db/query" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT name, phone_number, days_available FROM contacts WHERE days_available LIKE \"%Monday%\" LIMIT 5",
    "limit": 5
  }' | jq '.'

# Test 4: Get database schema
echo -e "\n4️⃣ Testing schema query..."
curl -s -X POST "$BASE_URL/db/schema" \
  -H "Content-Type: application/json" \
  -d '{"table_name": "contacts"}' | jq '.'

# Test 5: LLM tool execution - Count total contacts
echo -e "\n5️⃣ Testing LLM tool execution..."
curl -s -X POST "$BASE_URL/tools/single" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-1",
    "type": "tool_use",
    "name": "database_query",
    "input": {
      "query": "SELECT COUNT(*) as total FROM contacts"
    }
  }' | jq '.'

# Test 6: Search for contacts by name
echo -e "\n6️⃣ Testing contact search..."
curl -s -X POST "$BASE_URL/db/query" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT name, phone_number, email FROM contacts WHERE name LIKE \"%John%\" LIMIT 3",
    "limit": 3
  }' | jq '.'

# Test 7: Get contacts by availability
echo -e "\n7️⃣ Testing availability search..."
curl -s -X POST "$BASE_URL/db/query" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT name, days_available FROM contacts WHERE days_available LIKE \"%Friday%\" AND days_available LIKE \"%Saturday%\" LIMIT 3",
    "limit": 3
  }' | jq '.'

echo -e "\n✅ All API tests completed!"
echo -e "\n💡 Try these example queries:"
echo "   - Find contacts available on specific days"
echo "   - Search by name or phone number"
echo "   - Get contact statistics"
echo "   - Filter by location or other criteria"
