package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/lib/pq"              // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3"    // SQLite driver
)

// Connection manages database connections
type Connection struct {
	DB     *sql.DB
	Config *Config
}

// NewConnection creates a new database connection
func NewConnection(config *Config) (*Connection, error) {
	db, err := sql.Open(config.DriverName(), config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxConns)
	db.SetMaxIdleConns(config.MaxIdle)
	db.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if config.Type == "sqlite" {
		log.Printf("Connected to SQLite database: %s", config.FilePath)
	} else {
		log.Printf("Connected to PostgreSQL database: %s@%s:%d/%s", config.User, config.Host, config.Port, config.DBName)
	}

	return &Connection{
		DB:     db,
		Config: config,
	}, nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// Health checks if the database connection is healthy
func (c *Connection) Health() error {
	return c.DB.Ping()
}
