package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
)

// ConditionalAuthMiddleware applies authentication only to protected routes
func ConditionalAuthMiddleware(authClient authpb.AuthServiceClient, publicRoutes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for public routes
			if isPublicRoute(r.URL.Path, publicRoutes) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
				return
			}

			// Call gRPC `CheckAccess`
			_, err := authClient.CheckAccess(context.Background(), &authpb.CheckAccessRequest{Token: token})
			if err != nil {
				log.Println(err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isPublicRoute checks if a given path is in the public routes list
func isPublicRoute(path string, publicRoutes []string) bool {
	for _, route := range publicRoutes {
		if strings.HasPrefix(path, route) {
			return true
		}
	}
	return false
}
