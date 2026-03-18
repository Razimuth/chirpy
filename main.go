package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	// Initialize the configuration struct.
	apiCfg := &apiConfig{}

	// Create a new http.ServeMux
	mux := http.NewServeMux()

	// Register the healthz handler for the /healthz path
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	// Register the new /metrics and /reset handlers as methods on the apiConfig struct.
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidateChirp)

	// Wrap the file server with http.StripPrefix
	// http.StripPrefix removes the /app prefix from the request path before passing it to the file server
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	// Create an http.Server struct and set its Addr and Handler
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)

	// Use the server's ListenAndServe method to start the server
	if err := server.ListenAndServe(); err != nil {
		// Log a fatal error if the server fails to start
		log.Fatalf("Server error: %v", err)
	}
}
