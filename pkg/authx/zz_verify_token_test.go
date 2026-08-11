package authx

import (
	"os"
	"testing"
)

func TestVerifyTokenCLI(t *testing.T) {
	token, err := os.ReadFile("D:/Temp/opencode/token.txt")
	if err != nil {
		t.Fatal(err)
	}
	cfg := JWTConfig{Secret: "ride-hailing-dev-secret-change-me", Issuer: "ride-hailing", ExpireSeconds: 7200}
	m := NewManager(cfg)
	claims, err := m.Parse(string(token))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	t.Logf("parsed ok: user=%s role=%s jti=%s", claims.UserID, claims.Role, claims.JTI)
}