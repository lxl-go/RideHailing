package authx

import (
	"errors"
	"testing"
	"time"
)

func TestManagerGenerateAndParseBearer(t *testing.T) {
	manager := NewManager(JWTConfig{
		Secret:        "test-secret",
		Issuer:        "ride-hailing-test",
		ExpireSeconds: 3600,
	})

	token, err := manager.Generate(Claims{UserID: "1001", Role: "passenger"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if token.AccessToken == "" {
		t.Fatal("Generate() returned empty access token")
	}
	if token.TokenType != "Bearer" {
		t.Fatalf("TokenType = %q, want Bearer", token.TokenType)
	}
	if token.ExpiresIn != 3600 {
		t.Fatalf("ExpiresIn = %d, want 3600", token.ExpiresIn)
	}

	claims, err := manager.ParseBearer("Bearer " + token.AccessToken)
	if err != nil {
		t.Fatalf("ParseBearer() error = %v", err)
	}
	if claims.UserID != "1001" {
		t.Fatalf("UserID = %q, want 1001", claims.UserID)
	}
	if claims.Role != "passenger" {
		t.Fatalf("Role = %q, want passenger", claims.Role)
	}
	if claims.JTI == "" {
		t.Fatal("JTI is empty")
	}
}

func TestManagerRejectsMissingSecret(t *testing.T) {
	manager := NewManager(JWTConfig{Issuer: "ride-hailing-test"})

	_, err := manager.Generate(Claims{UserID: "1001", Role: "driver"})
	if !errors.Is(err, ErrMissingSecret) {
		t.Fatalf("Generate() error = %v, want ErrMissingSecret", err)
	}
}

func TestManagerRejectsExpiredToken(t *testing.T) {
	manager := NewManager(JWTConfig{
		Secret:        "test-secret",
		Issuer:        "ride-hailing-test",
		ExpireSeconds: -time.Hour.Milliseconds() / 1000,
	})

	token, err := manager.Generate(Claims{UserID: "1001", Role: "driver"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = manager.ParseBearer("Bearer " + token.AccessToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("ParseBearer() error = %v, want ErrInvalidToken", err)
	}
}

func TestManagerRejectsMalformedBearer(t *testing.T) {
	manager := NewManager(JWTConfig{Secret: "test-secret", Issuer: "ride-hailing-test"})

	_, err := manager.ParseBearer("bad-token")
	if !errors.Is(err, ErrMissingBearer) {
		t.Fatalf("ParseBearer() error = %v, want ErrMissingBearer", err)
	}
}
