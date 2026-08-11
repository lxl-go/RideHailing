package authx

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrMissingSecret = errors.New("jwt secret is required")
	ErrMissingBearer = errors.New("missing bearer token")
	ErrInvalidToken  = errors.New("invalid bearer token")
)

type JWTConfig struct {
	Secret        string `yaml:"secret"`
	Issuer        string `yaml:"issuer"`
	ExpireSeconds int64  `yaml:"expire_seconds"`
	Enabled       bool   `yaml:"enabled"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	JTI    string `json:"jti"`
}

type TokenPair struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type Manager struct {
	cfg JWTConfig
	now func() time.Time
}

type registeredClaims struct {
	UserID       string `json:"user_id"`
	LegacyUserID string `json:"userId"`
	UID          string `json:"uid"`
	Role         string `json:"role"`
	jwt.RegisteredClaims
}

func NewManager(cfg JWTConfig) *Manager {
	if strings.TrimSpace(cfg.Issuer) == "" {
		cfg.Issuer = "ride-hailing"
	}
	if cfg.ExpireSeconds == 0 {
		cfg.ExpireSeconds = 7200
	}
	return &Manager{cfg: cfg, now: time.Now}
}

func (m *Manager) Generate(claims Claims) (TokenPair, error) {
	if strings.TrimSpace(m.cfg.Secret) == "" {
		return TokenPair{}, ErrMissingSecret
	}
	now := m.now()
	jti := strings.TrimSpace(claims.JTI)
	if jti == "" {
		jti = fmt.Sprintf("%d", now.UnixNano())
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, registeredClaims{
		UserID: claims.UserID,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			Subject:   claims.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(m.cfg.ExpireSeconds) * time.Second)),
			ID:        jti,
		},
	})
	signed, err := token.SignedString([]byte(m.cfg.Secret))
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   m.cfg.ExpireSeconds,
	}, nil
}

func (m *Manager) ParseBearer(header string) (Claims, error) {
	tokenText := bearerToken(header)
	if tokenText == "" {
		return Claims{}, ErrMissingBearer
	}
	return m.Parse(tokenText)
}

func (m *Manager) Parse(tokenText string) (Claims, error) {
	if strings.TrimSpace(m.cfg.Secret) == "" {
		return Claims{}, ErrMissingSecret
	}
	rawClaims := &registeredClaims{}
	token, err := jwt.ParseWithClaims(tokenText, rawClaims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(m.cfg.Secret), nil
	}, jwt.WithIssuer(m.cfg.Issuer))
	if err != nil || !token.Valid {
		return Claims{}, ErrInvalidToken
	}
	userID := firstNonEmpty(rawClaims.UserID, rawClaims.LegacyUserID, rawClaims.UID, rawClaims.Subject)
	if userID == "" {
		return Claims{}, ErrInvalidToken
	}
	return Claims{
		UserID: userID,
		Role:   strings.TrimSpace(rawClaims.Role),
		JTI:    strings.TrimSpace(rawClaims.ID),
	}, nil
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
