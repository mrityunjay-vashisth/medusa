package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mrityunjay-vashisth/core-service/internal/handlers/utility"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"github.com/mrityunjay-vashisth/core-service/internal/services/authsvc"
	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
	"go.uber.org/zap"
)

// JWT secret key - should be loaded from env in production
var jwtKey = []byte("your-secure-jwt-secret-replace-in-production")

// AuthRequiredMiddleware creates a middleware that checks for valid auth token
func AuthRequiredMiddleware(serviceRegistry registry.ServiceRegistry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			log.Printf("Authorization Header: %s", authHeader)
			// token := extractTokenFromHeader(authHeader)

			if authHeader == "" {
				utility.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid Authorization header")
				return
			}

			// First try JWT verification (faster, doesn't require gRPC call)
			claims, err := validateToken(authHeader)
			if err == nil {
				// Create context with user claims
				ctx := context.WithValue(r.Context(), "claims", claims)
				ctx = context.WithValue(ctx, "username", claims.Username)
				ctx = context.WithValue(ctx, "role", claims.Role)
				ctx = context.WithValue(ctx, "tenantID", claims.TenantID)

				// Call next handler with enriched context
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Fallback to auth service verification via gRPC
			authService, ok := serviceRegistry.Get(registry.AuthService).(authsvc.Service)
			if !ok {
				utility.RespondWithError(w, http.StatusInternalServerError, "Internal auth service error")
				return
			}

			authClient := authService.GetClient()
			_, err = authClient.CheckAccess(r.Context(), &authpb.CheckAccessRequest{Token: authHeader})
			if err != nil {
				utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			// Continue with the request
			next.ServeHTTP(w, r)
		})
	}
}

// validateToken validates the JWT token and returns the claims
func validateToken(tokenString string) (*models.UserClaims, error) {
	claims := &models.UserClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check token expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// RoleRequiredMiddleware creates a middleware that checks for specific roles
func RoleRequiredMiddleware(requiredRoles []string, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context (added by AuthRequiredMiddleware)
			claims, ok := r.Context().Value("claims").(*models.UserClaims)
			if !ok {
				logger.Warn("No claims found in context, access denied")
				utility.RespondWithError(w, http.StatusForbidden, "Access denied")
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range requiredRoles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				logger.Warn("Insufficient permissions",
					zap.String("username", claims.Username),
					zap.String("role", claims.Role),
					zap.Strings("requiredRoles", requiredRoles))
				utility.RespondWithError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			// User has required role, proceed
			next.ServeHTTP(w, r)
		})
	}
}
