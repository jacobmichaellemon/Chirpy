package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

// Define your custom claims structure
type CustomClaims struct {
	jwt.RegisteredClaims
}

// Generate sample test data
func generateTestClaims(subject string, duration time.Duration) jwt.RegisteredClaims {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "chirpy-access",
		Subject:   subject,
	}
	return claims
}

func TestJWTClaims(t *testing.T) {
	// 1. Create Test Data
	claims := generateTestClaims("6ba7b810-9dad-11d1-80b4-00c04fd430c8", 1*time.Minute)
	secretKey := []byte("testsecret")

	// 2. Sign Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// 3. Test Parsing/Validation (Example verification)
	parsedToken, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil || !parsedToken.Valid {
		t.Errorf("Token should be valid: %v", err)
	}

	// 4. Assert Claims
	claimsObj := parsedToken.Claims.(*CustomClaims)
	if claimsObj.Subject != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Errorf("Expected subject 6ba7b810-9dad-11d1-80b4-00c04fd430c8, got %s", claimsObj.Subject)
	}
}

func TestImproperSignatureJWT(t *testing.T) {
	// 1. Setup
	realSecret := []byte("production-secret-key")
	attackerSecret := []byte("fake-secret-key")

	claims := &CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user_1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	// 2. The "Attacker" signs a token using their own fake secret
	attackerToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := attackerToken.SignedString(attackerSecret)
	if err != nil {
		t.Fatalf("Failed to sign attacker token: %v", err)
	}

	// 3. The "Server" tries to validate it using the REAL secret
	parsedToken, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		return realSecret, nil
	})

	// 4. Assertions
	// The library should return an error because the signature won't match
	if err == nil {
		t.Fatal("Security Failure: Accepted a token signed with the wrong secret!")
	}

	if !errors.Is(err, jwt.ErrSignatureInvalid) {
		t.Errorf("Expected error jwt.ErrSignatureInvalid, but got: %v", err)
	}

	if parsedToken != nil && parsedToken.Valid {
		t.Error("Token object marked as Valid despite signature mismatch")
	}
}

func TestExpiredJWTClaims(t *testing.T) {
	// 1. Create Test Data
	claims := generateTestClaims("6ba7b810-9dad-11d1-80b4-00c04fd430c8", 1*time.Microsecond)
	secretKey := []byte("testsecret")

	// 2. Sign Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// 3. Test Parsing/Validation (Example verification)
	parsedToken, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})

	time.Sleep(1 * time.Second) // wait out token, just in case

	if parsedToken.Valid {
		t.Errorf("Error: %v", err)
	}
}

func TestArgon2PasswordValidation(t *testing.T) {
	// 1. Setup
	password := "my-super-secret-123"
	wrongPassword := "not-the-secret"

	// We need a real Argon2id hash to test against.
	// You can generate this using your actual CreateHash helper.
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		t.Fatalf("Failed to create hash for test: %v", err)
	}

	// 2. Test Success: Correct Password
	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Errorf("CheckPasswordHash returned unexpected error: %v", err)
	}
	if !match {
		t.Error("Expected password to match hash, but it did not")
	}

	// 3. Test Failure: Incorrect Password
	match, err = CheckPasswordHash(wrongPassword, hash)
	if err != nil {
		t.Errorf("CheckPasswordHash returned unexpected error on mismatch: %v", err)
	}
	if match {
		t.Error("Expected password to NOT match hash, but it did")
	}

	// 4. Test Failure: Malformed Hash
	_, err = CheckPasswordHash(password, "invalid-hash-string")
	if err == nil {
		t.Error("Expected error when providing a malformed hash string")
	}
}
