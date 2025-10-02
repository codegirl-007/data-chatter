package main

import (
	"log"
	"net/http"
)

func main() {
	// Serve static files
	http.Handle("/", http.FileServer(http.Dir(".")))

	// Add CORS headers for API calls
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Proxy to the main API server
		http.Redirect(w, r, "http://localhost:8081"+r.URL.Path, http.StatusTemporaryRedirect)
	})

	port := "3000"
	log.Printf("🌐 Web interface starting on http://localhost:%s", port)
	log.Printf("📊 Make sure your API server is running on http://localhost:8081")
	log.Printf("🔗 Open http://localhost:%s in your browser", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start web server:", err)
	}
}
