package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggerMiddleware logs requests
func LoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            next.ServeHTTP(w, r)
            
			logger.Info("request",
                "method", r.Method,
                "path", r.URL.Path,
                "duration", time.Since(start).String(),
                "timestamp", time.Now().Format(time.RFC3339),
                "remote_addr", r.RemoteAddr,
                "forwarded_for", r.Header.Get("X-Forwarded-For"),
                "user_agent", r.UserAgent(),
            )
        })
    }
}
