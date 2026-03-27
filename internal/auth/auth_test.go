package auth

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()

	// Test success
	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	parsedID, err := ValidateJWT(token, secret)
	if err != nil || parsedID != userID {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Test expired
	expiredToken, _ := MakeJWT(userID, secret, -time.Hour)
	_, err = ValidateJWT(expiredToken, secret)
	if err == nil {
		t.Fatal("Expired token should have returned an error")
	}

	// Test wrong secret
	_, err = ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Fatal("Token validated with wrong secret")
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "secret", time.Hour)

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			tokenSecret: "secret",
			wantUserID:  userID,
			wantErr:     false,
		},
		{
			name:        "Invalid token",
			tokenString: "invalid.token.string",
			tokenSecret: "secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "Wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong_secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		headers       http.Header
		wantToken     string
		wantErr       bool
		expectedError error
	}{
		{
			name:      "Valid Bearer Token",
			headers:   http.Header{"Authorization": []string{"Bearer valid_token_string"}},
			wantToken: "valid_token_string",
			wantErr:   false,
		},
		{
			name:          "Missing Authorization Header",
			headers:       http.Header{},
			wantToken:     "",
			wantErr:       true,
			expectedError: errors.New("no authorization header found"),
		},
		{
			name:          "Invalid Format (No Bearer prefix)",
			headers:       http.Header{"Authorization": []string{"Token just_a_token"}},
			wantToken:     "",
			wantErr:       true,
			expectedError: errors.New("invalid authorization header format"),
		},
		{
			name:          "Invalid Format (Too many parts)",
			headers:       http.Header{"Authorization": []string{"Bearer one two three"}},
			wantToken:     "",
			wantErr:       true,
			expectedError: errors.New("invalid authorization header format"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(tt.headers)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.expectedError.Error() != err.Error() {
				t.Errorf("GetBearerToken() unexpected error = %v, want %v", err, tt.expectedError)
				return
			}
			if gotToken != tt.wantToken {
				t.Errorf("GetBearerToken() gotToken = %v, want %v", gotToken, tt.wantToken)
			}
		})
	}
}
