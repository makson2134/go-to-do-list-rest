package middleware

import (
	"context"
	"net/http"
	"strings"
	"to-do-list/internal/auth"

	"github.com/go-chi/render"
)

type CtxKey string

const UserIDKey CtxKey = "userID"


func AuthMiddleware(tokenManager *auth.TokenManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{"error": "Authorization header is required"})
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{"error": "Invalid Authorization header format"})
				return
			}

			tokenString := headerParts[1]

			userID, err := tokenManager.ValidateToken(tokenString)
			if err != nil {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{"error": "Invalid token"})
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
