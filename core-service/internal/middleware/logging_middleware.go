package middleware

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ContextKey is a custom type for storing values in context
type ContextKey string

const LoggerKey ContextKey = "logger"

// LoggingMiddleware logs HTTP requests and adds the logger to context
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Create a response writer to capture the status code
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Attach logger to context
			ctx := context.WithValue(r.Context(), LoggerKey, logger)
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(wrappedWriter, r)

			// Log request details
			logger.Info("HTTP Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", wrappedWriter.statusCode),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Duration("duration_ms", time.Since(startTime)),
			)
		})
	}
}

// responseWriter captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the response status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
