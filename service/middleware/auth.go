package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/provenance-templates/url-shortener/auth"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := auth.ExtractBearerToken(r)
		if !ok {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateToken(token)
		if err != nil {
			if errors.Is(err, auth.ErrUnauthorized) {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			http.Error(w, `{"error":"identity service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) string {
	if v, ok := r.Context().Value(UserIDKey).(string); ok {
		return v
	}
	return ""
}
