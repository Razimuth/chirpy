package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Razimuth/chirpy/internal/auth"
	"github.com/Razimuth/chirpy/internal/database"
	"github.com/google/uuid"
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

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	// Get the token from the Authorization header
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		// Return 401 Unauthorized if header is missing or malformed
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: missing or invalid token")
		return
	}

	// Validate the JWT and extract the user ID
	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		// Return 401 Unauthorized if token is invalid or expired
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: invalid token")
		return
	}

	// Decode the JSON body into the parameters struct
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters")
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedText := replaceProfaneWords(params.Body)

	// Save to database
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Body:      cleanedText, //params.Body,
		UserID:    userID,      //params.UserID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}
