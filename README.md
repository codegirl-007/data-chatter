# Data Chatter

A Go web server for secure database querying with LLM integration. Built with standard library only.

## Supported Databases

- **SQLite** (default)
- **PostgreSQL** 
- **MySQL** (newly added)

## Database Configuration

### SQLite (Default)
```bash
DB_TYPE=sqlite
DB_FILE_PATH=./contacts.db
```

### PostgreSQL
```bash
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=data_chatter
```

### MySQL
```bash
DB_TYPE=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_mysql_password
DB_NAME=contacts_db
```

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
│   │   └── database_handler.go    # Database-specific handlers
│   ├── tools/
│   │   └── database_tools.go       # Database query tools
│   ├── types/
│   │   └── tool_types.go          # Tool call data structures
│   └── middleware/
│       └── middleware.go          # HTTP middleware
├── go.mod                         # Go module file
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

### Using Make (if available):
```bash
make run
```

## API Endpoints

### Direct Database Access (Returns data directly)
- `POST /db/query` - Execute SQL SELECT queries
- `POST /db/schema` - Get database schema information

### LLM Tool Integration
- `GET /tools` - List available tools for LLM
- `POST /tools/execute` - Execute multiple tools (for LLM)
- `POST /tools/single` - Execute a single tool (for LLM)

### General
- `GET /` - Welcome message with API information
- `GET /health` - Health check endpoint
- `GET /api/*` - Generic API endpoint

## Features

- **🔒 Secure Database Access**: Read-only database queries with SQL injection protection
- **🤖 LLM Integration**: Tool call system for Claude integration
- **🛡️ Security**: SQL injection protection, query validation, and access controls
- **📊 Direct Data Access**: Database results returned directly to API (not to LLM)
- **🔍 Schema Discovery**: Automatic database schema inspection
- **⚡ Performance**: Connection pooling and optimized queries
- **🔄 Graceful shutdown** on SIGINT/SIGTERM
- **📝 Request logging** middleware
- **🌐 CORS support**
- **📋 JSON responses**
- **❤️ Health check** with uptime
- **📁 Standard Go project layout**

## Database Configuration

### SQLite (Default)
The server uses SQLite by default with a pre-populated contacts database:

```bash
# Default SQLite configuration (no setup required)
export DB_TYPE=sqlite
export DB_FILE=./contacts.db
```

### PostgreSQL (Optional)
To use PostgreSQL instead:

```bash
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=data_chatter
export DB_SSLMODE=disable
```

## Database Setup

The SQLite database is automatically created with 1000 sample contacts. To recreate it:

```bash
# Generate new sample data
go run scripts/setup_database.go

# Test database
go run scripts/test_database.go
```

## Available Tools

### Database Tools (for LLM)
- `database_query` - Execute SQL SELECT queries (schema is provided directly to LLM)

## Development

The server runs on port 8080 by default. You can test the endpoints:

```bash
# Health check
curl http://localhost:8080/health

# Direct database query (returns data directly)
curl -X POST http://localhost:8080/db/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT name, phone_number, days_available FROM contacts LIMIT 5",
    "limit": 10
  }'

# Get database schema
curl -X POST http://localhost:8080/db/schema \
  -H "Content-Type: application/json" \
  -d '{}'

# Get schema for specific table
curl -X POST http://localhost:8080/db/schema \
  -H "Content-Type: application/json" \
  -d '{"table_name": "contacts"}'

# List available tools for LLM
curl http://localhost:8080/tools

# Execute tool via LLM interface
curl -X POST http://localhost:8080/tools/single \
  -H "Content-Type: application/json" \
  -d '{
    "id": "1",
    "type": "tool_use",
    "name": "database_query",
    "input": {"query": "SELECT COUNT(*) as total FROM contacts"}
  }'
```
