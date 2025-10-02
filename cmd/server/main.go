// Package main provides the HTTP server for the data-chatter application.
// It handles LLM integration, database queries, and tool execution.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"data-chatter/internal/database"
	"data-chatter/internal/handlers"

	"github.com/joho/godotenv"
)

// main initializes the HTTP server with database connection, CORS middleware,
// and graceful shutdown handling.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	dbConfig := database.DefaultConfig()
	dbConn, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	handlers.InitializeToolEngine(dbConn)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsMiddleware(setupRoutes(dbConn)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("Server starting on :%s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}

// corsMiddleware provides Cross-Origin Resource Sharing support for web clients.
// It sets appropriate headers and handles preflight OPTIONS requests.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// setupRoutes configures all HTTP endpoints for the application.
// Returns a ServeMux with routes for health checks, LLM integration,
// database access, and tool execution.
func setupRoutes(dbConn *database.Connection) *http.ServeMux {
	mux := http.NewServeMux()

	dbHandler := handlers.NewDatabaseHandler(dbConn)
	llmHandler := handlers.NewLLMHandler(dbConn)

	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/llm/message", llmHandler.ProcessMessageHandler)
	mux.HandleFunc("/db/query", dbHandler.QueryHandler)
	mux.HandleFunc("/db/schema", dbHandler.SchemaHandler)
	mux.HandleFunc("/tools", handlers.ToolsHandler)
	mux.HandleFunc("/tools/execute", handlers.ToolCallHandler)
	mux.HandleFunc("/tools/single", handlers.SingleToolHandler)
	mux.HandleFunc("/api/", handlers.APIHandler)
	mux.HandleFunc("/", handlers.HomeHandler)

	return mux
}
