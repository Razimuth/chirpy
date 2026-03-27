package main

import (
	"net/http"
	"time"

	"github.com/Razimuth/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	// Extract token from "Authorization: Bearer <token>" header
	refreshToken := extractTokenFromHeader(r)

	// Look up the token in the database, checking expiration and revocation
	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken) // Use your SQL query/method
	if err != nil {
		// Token not found, expired, or revoked (handle specific errors)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate a new access token (expires in 1 hour)
	newAccessToken, err := auth.MakeJWT(user, cfg.jwtSecret, time.Hour)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Respond with the new access token
	respondWithJSON(w, http.StatusOK, UserResponse{
		//ID:           user.ID,
		//CreatedAt:    user.CreatedAt,
		//UpdatedAt:    user.UpdatedAt,
		//Email:        user.Email,
		Token: newAccessToken,
		//RefreshToken: refreshToken,
	})
}

//func RevokeHandler(w http.ResponseWriter, r *http.Request) {
// Extract token from "Authorization: Bearer <token>" header
//    refreshToken := extractTokenFromHeader(r)
//    if refreshToken == "" {
//        http.Error(w, "Unauthorized", http.StatusUnauthorized)
//        return
//    }
// Update the token record in the database
//    currentTime := time.Now()
// db.Exec(...) // Use your database driver to update the record:
// UPDATE refresh_tokens SET revoked_at = $1, updated_at = $2 WHERE token = $3
// Respond with 204 No Content status
//    w.WriteHeader(http.StatusNoContent)
//}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken := extractTokenFromHeader(r)
	//  Revoke the token in the database
	err := cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		// Handle cases where the token might not exist or other DB errors
		http.Error(w, "Unauthorized", http.StatusUnauthorized) // Or 500 depending on error
		return
	}

	// 2. Respond with a 204 No Content status
	w.WriteHeader(http.StatusNoContent)
}

// Helper function to extract token from header (used in refresh and revoke)
func extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

//userResponse := map[string]string{
//		"token": newAccessToken,
//	}

// ... (send JSON response with 200 status) ...
//}

// dbClient.GetUserFromRefreshToken implementation example (Go SQL pseudocode):
/*
func (db *DBClient) GetUserFromRefreshToken(token string) (uuid.UUID, error) {
    query := `
        SELECT user_id FROM refresh_tokens
        WHERE token = $1
        AND expires_at > CURRENT_TIMESTAMP
        AND revoked_at IS NULL
    `
    var userID uuid.UUID
    err := db.Pool.QueryRow(context.Background(), query, token).Scan(&userID)
    if err != nil {
        return uuid.Nil, fmt.Errorf("invalid or expired refresh token: %w", err)
    }
    return userID, nil
}
*/
