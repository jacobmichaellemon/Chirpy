package auth

import (
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
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
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
	headerValues := headers.Values("Authorization")
	if len(headerValues) == 0 {
		return "", nil //no auth headers to grab
	}

	token := headers.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		err := fmt.Errorf("No bearer token was found!")
		return "", err
	}
	return token, nil
}
