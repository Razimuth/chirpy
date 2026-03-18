package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Define the shape of the incoming JSON body
type ChirpRequest struct {
	Body string `json:"body"`
}

// Define the shape of a valid response body
//type ValidResponse struct {
//	Valid bool `json:"valid"`
//}

// Struct for the response body
type ChirpResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

// Define the shape of an error response body
type ErrorResponse struct {
	Error string `json:"error"`
}

// Helper function to respond with JSON data
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Marshal the payload into JSON bytes without extra whitespace
	response, err := json.Marshal(payload)
	if err != nil {
		// Fallback error if marshal fails
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error during JSON encoding"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Helper function to respond with an error message
func respondWithError(w http.ResponseWriter, code int, msg string) {
	// Use the generic JSON responder with the ErrorResponse struct
	respondWithJSON(w, code, ErrorResponse{Error: msg})
}

// Implementation for validating chirps
func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	// Decode the incoming JSON body
	decoder := json.NewDecoder(r.Body)
	var req ChirpRequest
	err := decoder.Decode(&req)
	if err != nil {
		// If the JSON is malformed or missing the "body" key
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format or missing 'body' field")
		return
	}
	// Validate the chirp length
	const maxChirpLength = 140
	if len(req.Body) > maxChirpLength {
		// Respond with HTTP 400 Bad Request if too long
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedText := replaceProfaneWords(req.Body)

	respondWithJSON(w, http.StatusOK, ChirpResponse{CleanedBody: cleanedText})
	// If valid, respond with HTTP 200 OK and the success body
	//		respondWithJSON(w, http.StatusOK, ValidResponse{Valid: true})
	//	}
}

// replaceProfaneWords cleans the input string by replacing any occurrences
// of the target bad words with '****', case-insensitively.
func replaceProfaneWords(text string) string {
	profaneWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	// Split the text into words based on spaces
	words := strings.Fields(text)
	cleanedWords := []string{}

	for _, word := range words {
		// Create a lowercase version for comparison
		lowerWord := strings.ToLower(word)

		// Check if the word exactly matches one of the profane words
		if profaneWords[lowerWord] {
			cleanedWords = append(cleanedWords, "****")
		} else {
			cleanedWords = append(cleanedWords, word)
		}
	}

	// Rejoin the words into a single string
	return strings.Join(cleanedWords, " ")
}
