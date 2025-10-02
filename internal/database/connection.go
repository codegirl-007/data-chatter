// Package database provides database connection management and configuration
// for SQLite, PostgreSQL, and MySQL databases.
package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Connection represents an active database connection with configuration.
type Connection struct {
	DB     *sql.DB
	Config *Config
}

// NewConnection establishes a new database connection using the provided configuration.
// It configures connection pooling, tests the connection, and logs the successful connection.
func NewConnection(config *Config) (*Connection, error) {
	db, err := sql.Open(config.DriverName(), config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(config.MaxConns)
	db.SetMaxIdleConns(config.MaxIdle)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if config.Type == "sqlite" {
		log.Printf("Connected to SQLite database: %s", config.FilePath)
	} else {
		log.Printf("Connected to %s database: %s@%s:%d/%s", config.Type, config.User, config.Host, config.Port, config.DBName)
	}

	return &Connection{
		DB:     db,
		Config: config,
	}, nil
}

// Close terminates the database connection and releases associated resources.
func (c *Connection) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// Health performs a ping test to verify the database connection is still active.
func (c *Connection) Health() error {
	return c.DB.Ping()
}
