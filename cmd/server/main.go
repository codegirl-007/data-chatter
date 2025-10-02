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

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Initialize database connection
	dbConfig := database.DefaultConfig()
	dbConn, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Initialize tool engine
	handlers.InitializeToolEngine(dbConn)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Create a new HTTP server with CORS middleware
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsMiddleware(setupRoutes(dbConn)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("Server starting on :%s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Server shutting down...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}

func setupRoutes(dbConn *database.Connection) *http.ServeMux {
	mux := http.NewServeMux()

	// Initialize handlers
	dbHandler := handlers.NewDatabaseHandler(dbConn)
	llmHandler := handlers.NewLLMHandler(dbConn)

	// Health check endpoint
	mux.HandleFunc("/health", handlers.HealthHandler)

	// LLM integration endpoint
	mux.HandleFunc("/llm/message", llmHandler.ProcessMessageHandler)

	// Database endpoints (direct data access)
	mux.HandleFunc("/db/query", dbHandler.QueryHandler)
	mux.HandleFunc("/db/schema", dbHandler.SchemaHandler)

	// Tool endpoints (for LLM integration)
	mux.HandleFunc("/tools", handlers.ToolsHandler)
	mux.HandleFunc("/tools/execute", handlers.ToolCallHandler)
	mux.HandleFunc("/tools/single", handlers.SingleToolHandler)

	// API routes
	mux.HandleFunc("/api/", handlers.APIHandler)

	// Root endpoint
	mux.HandleFunc("/", handlers.HomeHandler)

	return mux
}
