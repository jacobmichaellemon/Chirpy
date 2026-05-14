package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	return hash, err
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	return match, err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// Create the Claims
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "chirpy-access",
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return ss, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var claims jwt.RegisteredClaims
	// ParseWithClaims parses the token and populates the 'claims' struct
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		// Important: Always validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(tokenSecret), nil
	})

	var nilID uuid.NullUUID = uuid.NullUUID{}

	// Check for parsing or validation errors
	if err != nil {
		return nilID.UUID, err
	}

	// Verify if the token is actually valid (signatures and standard claims)
	if !token.Valid {
		return nilID.UUID, jwt.ErrSignatureInvalid
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nilID.UUID, err
	}

	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		err := fmt.Errorf("No bearer token was found!")
		return "", err
	}
	return token, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	apiKey := headers.Get("Authorization")
	apiKey = strings.TrimPrefix(apiKey, "ApiKey ")
	if apiKey == "" {
		err := fmt.Errorf("No api key was found!")
		return "", err
	}
	return apiKey, nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	encodeString := hex.EncodeToString(key)
	return encodeString
}
