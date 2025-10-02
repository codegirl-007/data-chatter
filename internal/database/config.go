package database

import (
	"fmt"
	"os"
	"strconv"
)

// Config contains database connection parameters and connection pool settings.
type Config struct {
	Type     string // Database type: "postgres", "sqlite", or "mysql"
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	MaxIdle  int
	FilePath string // For SQLite file path
}

// DefaultConfig creates a database configuration from environment variables.
// Defaults to SQLite if DB_TYPE is not set, otherwise configures based on DB_TYPE.
func DefaultConfig() *Config {
	dbType := getEnv("DB_TYPE", "sqlite")

	if dbType == "sqlite" {
		return &Config{
			Type:     "sqlite",
			FilePath: getEnv("DB_FILE", "./contacts.db"),
			MaxConns: getEnvInt("DB_MAX_CONNS", 10),
			MaxIdle:  getEnvInt("DB_MAX_IDLE", 5),
		}
	}

	if dbType == "mysql" {
		return &Config{
			Type:     "mysql",
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 3306),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "data_chatter"),
			MaxConns: getEnvInt("DB_MAX_CONNS", 10),
			MaxIdle:  getEnvInt("DB_MAX_IDLE", 5),
		}
	}

	return &Config{
		Type:     "postgres",
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "data_chatter"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
		MaxConns: getEnvInt("DB_MAX_CONNS", 10),
		MaxIdle:  getEnvInt("DB_MAX_IDLE", 5),
	}
}

// ConnectionString generates the appropriate connection string for the database type.
func (c *Config) ConnectionString() string {
	if c.Type == "sqlite" {
		return c.FilePath
	}

	if c.Type == "mysql" {
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.User, c.Password, c.Host, c.Port, c.DBName)
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// DriverName returns the database driver name for the configured database type.
func (c *Config) DriverName() string {
	if c.Type == "sqlite" {
		return "sqlite3"
	}
	if c.Type == "mysql" {
		return "mysql"
	}
	return "postgres"
}

// getEnv retrieves an environment variable with a fallback default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an environment variable as an integer with a fallback default value.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
