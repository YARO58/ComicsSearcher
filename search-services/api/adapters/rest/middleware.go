package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"yadro.com/course/api/core"
)

type loginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func WithAuth(auth core.Authenticator, log *slog.Logger) core.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if err := auth.ValidateToken(parts[1]); err != nil {
				log.Error("invalid token", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func WithRateLimit(limiter core.RateLimiter, log *slog.Logger) core.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limiter.Wait()
			next.ServeHTTP(w, r)
		})
	}
}

func WithConcurrencyLimit(limiter core.ConcurrencyLimiter, log *slog.Logger) core.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Acquire() {
				w.WriteHeader(http.StatusServiceUnavailable)
				log.Debug("concurrency limit exceeded")
				return
			}
			defer limiter.Release()
			next.ServeHTTP(w, r)
			log.Debug("concurrency limit released")
		})
	}
}

func NewLoginHandler(auth core.Authenticator, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode login request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := auth.GenerateToken(req.Name, req.Password)
		if err != nil {
			log.Error("failed to generate token", "error", err)
			log.Info("admin user: %s, admin password: %s", req.Name, req.Password)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		_, err = w.Write([]byte(token))
		if err != nil {
			log.Error("failed to write token", "error", err)
		}
	}
}
