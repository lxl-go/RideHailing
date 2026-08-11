package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"ride-hailing/pkg/httpx"
)

const (
	HeaderRequestID  = "X-Request-Id"
	claimsContextKey = "jwt_claims"
)

// ============================================================
// CORS
// ============================================================

type CORSConfig struct {
	AllowOrigins     []string      `mapstructure:"allow_origins" yaml:"allow_origins"`
	AllowMethods     []string      `mapstructure:"allow_methods" yaml:"allow_methods"`
	AllowHeaders     []string      `mapstructure:"allow_headers" yaml:"allow_headers"`
	ExposeHeaders    []string      `mapstructure:"expose_headers" yaml:"expose_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age" yaml:"max_age"`
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization", HeaderRequestID},
		ExposeHeaders: []string{HeaderRequestID},
		MaxAge:        12 * time.Hour,
	}
}

func CORS(cfg CORSConfig) gin.HandlerFunc {
	cfg = fillCORSDefaults(cfg)
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		if !originAllowed(origin, cfg.AllowOrigins) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if allowAllOrigins(cfg.AllowOrigins) && !cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
		c.Header("Access-Control-Max-Age", strconv.Itoa(int(cfg.MaxAge.Seconds())))
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func fillCORSDefaults(cfg CORSConfig) CORSConfig {
	def := DefaultCORSConfig()
	if len(cfg.AllowOrigins) == 0 {
		cfg.AllowOrigins = def.AllowOrigins
	}
	if len(cfg.AllowMethods) == 0 {
		cfg.AllowMethods = def.AllowMethods
	}
	if len(cfg.AllowHeaders) == 0 {
		cfg.AllowHeaders = def.AllowHeaders
	}
	if len(cfg.ExposeHeaders) == 0 {
		cfg.ExposeHeaders = def.ExposeHeaders
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = def.MaxAge
	}
	return cfg
}

func originAllowed(origin string, allowOrigins []string) bool {
	for _, allowed := range allowOrigins {
		if allowed == "*" || strings.EqualFold(allowed, origin) {
			return true
		}
	}
	return false
}

func allowAllOrigins(allowOrigins []string) bool {
	for _, allowed := range allowOrigins {
		if allowed == "*" {
			return true
		}
	}
	return false
}

// ============================================================
// ServiceToken
// ============================================================

type ServiceTokenConfig struct {
	Token string `mapstructure:"token" yaml:"token"`
}

func ServiceTokenAuth(cfg ServiceTokenConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Service-Token")
		if token == "" {
			abortUnauthorized(c, "missing_service_token", "missing X-Service-Token header")
			return
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(cfg.Token)) != 1 {
			abortUnauthorized(c, "invalid_service_token", "invalid X-Service-Token")
			return
		}
		c.Next()
	}
}

// ============================================================
// JWT
// ============================================================

type JWTConfig struct {
	Secret        string `mapstructure:"secret" yaml:"secret"`
	Issuer        string `mapstructure:"issuer" yaml:"issuer"`
	ExpireSeconds int64  `mapstructure:"expire_seconds" yaml:"expire_seconds"`
}

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	JTI    string `json:"jti"`
}

type registeredJWTClaims struct {
	UserID       string `json:"user_id"`
	LegacyUserID string `json:"userId"`
	UID          string `json:"uid"`
	Role         string `json:"role"`
	jwt.RegisteredClaims
}

func effectiveJTI(claims *registeredJWTClaims) string {
	if claims == nil {
		return ""
	}
	return strings.TrimSpace(claims.RegisteredClaims.ID)
}

func effectiveUserID(claims *registeredJWTClaims) string {
	if claims == nil {
		return ""
	}
	for _, value := range []string{
		claims.UserID,
		claims.LegacyUserID,
		claims.UID,
		claims.RegisteredClaims.Subject,
	} {
		if uid := strings.TrimSpace(value); uid != "" {
			return uid
		}
	}
	return ""
}

type JWT struct {
	cfg JWTConfig
}

func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		Issuer:        "xiaolanshu",
		ExpireSeconds: 7200,
	}
}

func NewJWT(cfg JWTConfig) *JWT {
	def := DefaultJWTConfig()
	if cfg.Issuer == "" {
		cfg.Issuer = def.Issuer
	}
	if cfg.ExpireSeconds <= 0 {
		cfg.ExpireSeconds = def.ExpireSeconds
	}
	return &JWT{cfg: cfg}
}

func (j *JWT) GenerateToken(claims JWTClaims) (string, error) {
	if strings.TrimSpace(j.cfg.Secret) == "" {
		return "", errors.New("jwt secret is required")
	}
	now := time.Now()
	jti := claims.JTI
	if strings.TrimSpace(jti) == "" {
		jti = fmt.Sprintf("%d", now.UnixNano())
		claims.JTI = jti
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, registeredJWTClaims{
		UserID: claims.UserID,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.cfg.Issuer,
			Subject:   claims.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(j.cfg.ExpireSeconds) * time.Second)),
			ID:        jti,
		},
	})
	return token.SignedString([]byte(j.cfg.Secret))
}

func (j *JWT) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(j.cfg.Secret) == "" {
			abortUnauthorized(c, "jwt_secret_missing", "jwt secret is not configured")
			return
		}
		tokenText := bearerToken(c.GetHeader("Authorization"))
		if tokenText == "" {
			abortUnauthorized(c, "missing_token", "missing bearer token")
			return
		}
		claims := &registeredJWTClaims{}
		token, err := jwt.ParseWithClaims(tokenText, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(j.cfg.Secret), nil
		}, jwt.WithIssuer(j.cfg.Issuer))
		if err != nil || !token.Valid {
			abortUnauthorized(c, "invalid_token", "invalid bearer token")
			return
		}
		userID := effectiveUserID(claims)
		if userID == "" {
			abortUnauthorized(c, "invalid_token", "invalid bearer token")
			return
		}
		c.Set(claimsContextKey, JWTClaims{
			UserID: userID,
			Role:   claims.Role,
			JTI:    effectiveJTI(claims),
		})
		c.Next()
	}
}

func (j *JWT) ExpireSeconds() int64 {
	return j.cfg.ExpireSeconds
}

func JWTAuth(cfg JWTConfig) gin.HandlerFunc {
	return NewJWT(cfg).Auth()
}

func JWTAuthWithBlackList(jwtCfg JWTConfig, redisCli *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(jwtCfg.Secret) == "" {
			abortUnauthorized(c, "jwt_secret_missing", "jwt secret is not configured")
			return
		}
		tokenText := bearerToken(c.GetHeader("Authorization"))
		if tokenText == "" {
			abortUnauthorized(c, "missing_token", "未携带登录凭证，请重新登录")
			return
		}
		claims := &registeredJWTClaims{}
		token, err := jwt.ParseWithClaims(tokenText, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(jwtCfg.Secret), nil
		}, jwt.WithIssuer(jwtCfg.Issuer))
		if err != nil || !token.Valid {
			abortUnauthorized(c, "invalid_token", "登录凭证无效或已过期")
			return
		}
		jti := effectiveJTI(claims)
		if jti == "" {
			abortUnauthorized(c, "invalid_token", "登录凭证无效或已过期")
			return
		}
		ctx := c.Request.Context()
		blackKey := fmt.Sprintf("jwt:black:jti:%s", jti)
		exists, err := redisCli.Exists(ctx, blackKey).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": gin.H{"code": "auth_unavailable", "message": "认证服务暂不可用，请稍后重试"},
			})
			return
		}
		if exists > 0 {
			abortUnauthorized(c, "token_revoked", "该登录凭证已失效，请重新登录")
			return
		}
		userID := effectiveUserID(claims)
		if userID == "" {
			abortUnauthorized(c, "invalid_token", "登录凭证无效或已过期")
			return
		}
		c.Set(claimsContextKey, JWTClaims{
			UserID: userID,
			Role:   claims.Role,
			JTI:    jti,
		})
		c.Next()
	}
}

func ClaimsFromContext(c *gin.Context) JWTClaims {
	value, ok := c.Get(claimsContextKey)
	if !ok {
		return JWTClaims{}
	}
	claims, _ := value.(JWTClaims)
	return claims
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func abortUnauthorized(c *gin.Context, code string, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{"code": code, "message": message},
	})
}

// ============================================================
// RequestID
// ============================================================

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" {
			id = uuid.New().String()
		}
		c.Set("request_id", id)
		c.Header(HeaderRequestID, id)
		c.Next()
	}
}

// ============================================================
// RateLimiter
// ============================================================

func RateLimiter(redisCli interface{}, maxRequests int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// ============================================================
// GlobalError
// ============================================================

func GlobalError() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			httpx.Error(c, c.Errors.Last().Err)
		}
	}
}

// ============================================================
// ContextPropagation — inject X-User-Id / X-Request-Id into context
// ============================================================

func ContextPropagation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if uid := c.GetHeader("X-User-Id"); uid != "" {
			ctx = context.WithValue(ctx, "x-user-id", uid)
			if id, err := strconv.ParseInt(uid, 10, 64); err == nil {
				ctx = context.WithValue(ctx, "user_id", id)
			}
		}
		if rid := c.GetHeader(HeaderRequestID); rid != "" {
			ctx = context.WithValue(ctx, "x-request-id", rid)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// OutgoingContext returns the request context with propagated values.
func OutgoingContext(c *gin.Context) context.Context {
	return c.Request.Context()
}

// GetUserID extracts user_id from context (set by ContextPropagation).
func GetUserID(c *gin.Context) (int64, error) {
	uid := c.GetInt64("user_id")
	if uid == 0 {
		return 0, fmt.Errorf("missing user id")
	}
	return uid, nil
}

// GetRequestID returns the request ID from context.
func GetRequestID(c *gin.Context) string {
	if id, ok := c.Get("request_id"); ok {
		return fmt.Sprintf("%v", id)
	}
	return ""
}
