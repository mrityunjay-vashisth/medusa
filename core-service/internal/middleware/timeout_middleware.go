package middleware

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// TimeoutMiddleware adds a timeout to the request context
func TimeoutMiddleware(timeout time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to handle timeouts
			done := make(chan struct{})

			// Execute the request with the timeout context
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Request completed normally
				return
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					// Log the timeout
					logger := r.Context().Value("logger")
					if zapLogger, ok := logger.(*zap.Logger); ok {
						zapLogger.Warn("Request timeout exceeded",
							zap.String("path", r.URL.Path),
							zap.Duration("timeout", timeout))
					}

					// Return timeout error to client
					w.WriteHeader(http.StatusRequestTimeout)
					w.Write([]byte(`{"error": "Request timeout exceeded"}`))
					return
				}
			}
		}
	}
}
