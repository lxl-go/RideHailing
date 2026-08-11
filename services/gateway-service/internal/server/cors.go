package server

import (
	"net/http"
	"net/url"
	"strconv"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

const (
	corsAllowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsAllowHeaders = "Content-Type, Authorization, X-Requested-With, X-Token, X-User-Id, X-Trace-Id, Idempotency-Key"
	corsMaxAge       = "86400"
)

func NewCORSFilter() khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setCORSHeaders(w, r)
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if isAllowedLocalH5Origin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Vary", "Origin")
	}
	w.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)
	w.Header().Set("Access-Control-Allow-Headers", corsAllowHeaders)
	w.Header().Set("Access-Control-Max-Age", corsMaxAge)
}

func isAllowedLocalH5Origin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "http" {
		return false
	}

	host := u.Hostname()
	if host != "localhost" && host != "127.0.0.1" {
		return false
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return false
	}
	return port >= 5173 && port <= 5179
}
