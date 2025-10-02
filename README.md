# Data Chatter

A Go web server for secure database querying with LLM integration. Built with standard library only.

## Tool Call Architecture

This system uses a **simplified tool call architecture** with only one tool: `database_query`.

### How Tool Calls Work

1. **LLM Integration Flow:**
   ```
   User Message → LLM (with schema) → Tool Call → Database → Results → User
   ```

2. **Tool Execution Process:**
   - LLM receives database schema directly in system prompt
     - **Code:** `internal/llm/anthropic_client.go:getDatabaseSchema()`
   - LLM constructs SQL query and calls `database_query` tool
     - **Code:** `internal/llm/anthropic_client.go:getAvailableTools()`
   - Tool validates query (SELECT only, security checks)
     - **Code:** `internal/tools/database_tools.go:Validate()`
   - Database executes query and returns results
     - **Code:** `internal/tools/database_tools.go:Execute()`
   - Results sent directly to user (not back to LLM)
     - **Code:** `internal/handlers/llm_handler.go:ProcessMessageHandler()`

3. **HTTP-Based Tool Execution:**
   - LLM handler makes HTTP POST to `/tools/single`
     - **Code:** `internal/handlers/llm_handler.go:executeToolCall()`
   - SingleToolHandler processes the tool call
     - **Code:** `internal/handlers/handlers.go:SingleToolHandler()`
   - ToolEngine executes the database query
     - **Code:** `internal/engine/tool_engine.go:ExecuteTool()`
   - Results returned as JSON
     - **Code:** `internal/tools/database_tools.go:Execute()`

### Available Tools

#### Database Tools (for LLM)
- `database_query` - Execute SQL SELECT queries (schema provided directly to LLM)

**Tool Definition:**
```json
{
  "name": "database_query",
  "description": "Execute a read-only SQL SELECT query on the database",
  "input_schema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "SQL SELECT query to execute (include LIMIT clause if needed)"
      }
    },
    "required": ["query"]
  }
}
```

### Tool Call Security

- **Read-only queries only** - Only SELECT statements allowed
- **SQL injection protection** - Query validation and sanitization
- **Dangerous keyword blocking** - Prevents DROP, DELETE, UPDATE, etc.
- **No data exposure to LLM** - Results go directly to user

## Supported Databases

- **SQLite** (default)
- **PostgreSQL** 
- **MySQL**

## Database Configuration

Database configuration is handled in `internal/database/config.go:DefaultConfig()`.

### SQLite (Default)
```bash
DB_TYPE=sqlite
DB_FILE_PATH=./contacts.db
```
- **Connection:** `internal/database/connection.go:NewConnection()`
- **Driver:** `github.com/mattn/go-sqlite3`

### PostgreSQL
```bash
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=data_chatter
```
- **Connection:** `internal/database/connection.go:NewConnection()`
- **Driver:** `github.com/lib/pq`

### MySQL
```bash
DB_TYPE=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_mysql_password
DB_NAME=contacts_db
```
- **Connection:** `internal/database/connection.go:NewConnection()`
- **Driver:** `github.com/go-sql-driver/mysql`

## Project Structure

```
data-chatter/
├── cmd/server/main.go              # Application entry point
├── internal/
│   ├── database/
│   │   ├── config.go              # Database configuration
│   │   └── connection.go           # Database connection management
│   ├── engine/
│   │   └── tool_engine.go         # Tool execution engine
│   ├── handlers/
│   │   ├── handlers.go            # HTTP handlers
│   │   ├── database_handler.go    # Database-specific handlers
│   │   └── llm_handler.go         # LLM integration handler
│   ├── llm/
│   │   └── anthropic_client.go    # Anthropic API client
│   ├── tools/
│   │   └── database_tools.go      # Database query tools
│   ├── types/
│   │   └── tool_types.go          # Tool call data structures
│   └── middleware/
│       └── middleware.go          # HTTP middleware
├── web/                           # Web UI
│   ├── index.html                 # Web interface
│   ├── server.go                  # Web server
│   └── README.md                  # Web UI documentation
├── scripts/                       # Utility scripts
│   ├── start_full_stack.sh        # Start both API and web servers
│   ├── test_api.go                # API testing script
│   └── test_curl.sh               # cURL testing script
├── .env.example                   # Environment variables template
├── go.mod                         # Go module file
├── go.sum                         # Go module checksums
└── README.md                      # This file
```

## Running the Server

### Using Go directly:
```bash
go run cmd/server/main.go
```

### Building and running:
```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

## API Endpoints

### LLM Integration
- `POST /llm/message` - Send message to LLM with tool execution
  - **Handler:** `internal/handlers/llm_handler.go:ProcessMessageHandler()`

### Direct Database Access (Returns data directly)
- `POST /db/query` - Execute SQL SELECT queries
  - **Handler:** `internal/handlers/database_handler.go:QueryHandler()`
- `GET /db/schema` - Get database schema information (redirects to LLM integration)
  - **Handler:** `internal/handlers/database_handler.go:SchemaHandler()`

### Tool Integration (for LLM)
- `GET /tools` - List available tools for LLM
  - **Handler:** `internal/handlers/handlers.go:ToolsHandler()`
- `POST /tools/execute` - Execute multiple tools (for LLM)
  - **Handler:** `internal/handlers/handlers.go:ToolCallHandler()`
- `POST /tools/single` - Execute a single tool (for LLM)
  - **Handler:** `internal/handlers/handlers.go:SingleToolHandler()`

### General
- `GET /` - Welcome message with API information
  - **Handler:** `internal/handlers/handlers.go:HomeHandler()`
- `GET /health` - Health check endpoint
  - **Handler:** `internal/handlers/handlers.go:HealthHandler()`
- `GET /api/*` - Generic API endpoint
  - **Handler:** `internal/handlers/handlers.go:APIHandler()`

## Features

- **🔒 Secure Database Access**: Read-only database queries with SQL injection protection
  - **Code:** `internal/tools/database_tools.go:Validate()`
- **🤖 LLM Integration**: Tool call system for Claude integration
  - **Code:** `internal/llm/anthropic_client.go:ProcessMessage()`
- **🛡️ Security**: SQL injection protection, query validation, and access controls
  - **Code:** `internal/tools/database_tools.go:Validate()`
- **📊 Direct Data Access**: Database results returned directly to API (not to LLM)
  - **Code:** `internal/handlers/llm_handler.go:ProcessMessageHandler()`
- **🔍 Schema Discovery**: Automatic database schema inspection
  - **Code:** `internal/llm/anthropic_client.go:getDatabaseSchema()`
- **⚡ Performance**: Connection pooling and optimized queries
  - **Code:** `internal/database/connection.go:NewConnection()`
- **🔄 Graceful shutdown** on SIGINT/SIGTERM
  - **Code:** `cmd/server/main.go:main()`
- **📝 Request logging** middleware
  - **Code:** `internal/middleware/middleware.go`
- **🌐 CORS support**
  - **Code:** `cmd/server/main.go:corsMiddleware()`
- **📋 JSON responses**
  - **Code:** All handlers in `internal/handlers/`
- **❤️ Health check** with uptime
  - **Code:** `internal/handlers/handlers.go:HealthHandler()`
- **📁 Standard Go project layout**
  - **Structure:** See Project Structure section above

## Development

The server runs on port 8081 by default. You can test the endpoints:

```bash
# Health check
curl http://localhost:8081/health

# LLM integration (requires ANTHROPIC_API_KEY)
curl -X POST http://localhost:8081/llm/message \
  -H "Content-Type: application/json" \
  -d '{"message": "fetch me all contacts available on Monday"}'

# Direct database query (returns data directly)
curl -X POST http://localhost:8081/db/query \
  -H "Content-Type: application/json" \
  -d '{"query": "SELECT name, phone_number, days_available FROM contacts LIMIT 5"}'

# Get database schema (now redirects to LLM integration)
curl http://localhost:8081/db/schema

# List available tools for LLM
curl http://localhost:8081/tools

# Execute tool via LLM interface
curl -X POST http://localhost:8081/tools/single \
  -H "Content-Type: application/json" \
  -d '{
    "id": "1",
    "type": "tool_use",
    "name": "database_query",
    "input": {"query": "SELECT COUNT(*) as total FROM contacts"}
  }'
```

## Environment Variables

Create a `.env` file with:

```bash
# Anthropic API Configuration
ANTHROPIC_API_KEY=your_anthropic_api_key_here

# Database Configuration
DB_TYPE=sqlite
DB_FILE_PATH=./contacts.db

# Server Configuration
PORT=8081
```

## Web UI

A simple web interface is available in the `web/` directory:

```bash
# Start web server
cd web && go run server.go

# Or use the full stack script
./scripts/start_full_stack.sh
```

Access the web UI at `http://localhost:3000` to interact with the API through a friendly interface.