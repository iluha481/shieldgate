package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", duration.String(),
			"remote_addr", r.RemoteAddr,
		)
	})
}
