// handler.go
package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/Razimuth/chirpy/internal/database"
)

// apiConfig holds all stateful, in-memory data.
type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
}

// middlewareMetricsInc is a middleware that increments the fileserverHits counter
// every time a request passes through it.
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the atomic.Int32's .Add() method to safely increment the counter.
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// handlerMetrics writes the current number of requests as plain text to the HTTP response.
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	hits := cfg.fileserverHits.Load()
	htmlResponce := fmt.Sprintf(`
<html>
<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>
</html>`, hits)

	w.Write([]byte(htmlResponce))
}

// handlerReset resets the fileserverHits counter back to 0.
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "This endpoint is only available in the development environment")
		return
	}
	cfg.fileserverHits.Store(0)
	//	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//	w.WriteHeader(http.StatusOK)
	//	w.Write([]byte("Hits reset to 0"))
	//}

	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Could not delete all users: %v", err))
		return
	}

	//	respondWithJSON(w, http.StatusOK, map[string]string{"status": "Hits reset to 0 and all users deleted"})
	respondWithJSON(w, http.StatusOK, "Hits reset to 0 and all users deleted")
}

// healthzHandler handles requests to the /healthz endpoint
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// Write the 200 OK status code. This must be called before w.Write to set the status correctly.
	w.WriteHeader(http.StatusOK) // http.StatusOK is a constant for 200
	// Write the body text "OK"
	w.Write([]byte("OK"))
}
