package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/marathozin/notes-api-go/internal/service"
	"github.com/marathozin/notes-api-go/pkg/response"
)

type contextKey string

const UserIDKey contextKey = "userID"

// Auth проверяет Bearer-токен и кладёт userID в контекст запроса.
func Auth(ts *service.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			userID, err := ts.ValidateAccess(token)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID извлекает userID из контекста (после прохождения Auth middleware).
func GetUserID(r *http.Request) int64 {
	id, _ := r.Context().Value(UserIDKey).(int64)
	return id
}
