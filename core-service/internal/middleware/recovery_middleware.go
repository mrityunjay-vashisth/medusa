package middleware

import (
	"context"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

// RecoveryMiddleware catches panics and recovers, logging the error
func RecoveryMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					stack := debug.Stack()
					logger.Error("Panic recovered in HTTP handler",
						zap.Any("error", err),
						zap.String("stack", string(stack)),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
						zap.String("remote_addr", r.RemoteAddr),
					)

					// Return a 500 Internal Server Error response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "Internal server error"}`))
				}
			}()

			// Add logger to the request context for use in handlers
			ctx := r.Context()
			ctx = context.WithValue(ctx, "logger", logger)

			// Process the request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
