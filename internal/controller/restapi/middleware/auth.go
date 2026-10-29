package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/iluha481/shieldgate/pkg/jwt"
)

type contextKey string

const userIDCtxKey contextKey = "user_id"

func GetUserID(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDCtxKey).(int)
	return id, ok
}

func Auth(jwtManager *jwt.Manager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"authorization header required"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"invalid format, expected: Bearer <token>"}`, http.StatusUnauthorized)
			return
		}

		sub, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		userID, err := strconv.Atoi(sub)
		if err != nil {
			http.Error(w, `{"error":"invalid subject in token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDCtxKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
