package apiserver

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/handlers/adminhdlr"
	"github.com/mrityunjay-vashisth/core-service/internal/handlers/authhdlr"
	"github.com/mrityunjay-vashisth/core-service/internal/handlers/onboardinghdlr"
	"github.com/mrityunjay-vashisth/core-service/internal/middleware"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"github.com/mrityunjay-vashisth/go-apigen/pkg/generator"
	"go.uber.org/zap"
)

// Constants for OpenAPI spec file paths
const (
	openapiDir     = "../internal/config/openapi"
	authSpecFile   = "auth.yaml"
	adminSpecFile  = "admin.yaml"
	tenantSpecFile = "onboarding.yaml"
)

// APIServer holds the router and related components
type APIServer struct {
	Router   *mux.Router
	Logger   *zap.Logger
	Registry registry.ServiceRegistry
}

// NewAPIServer initializes the API server with all routers
func NewAPIServer(ctx context.Context, db db.DBClientInterface, serviceRegistry registry.ServiceRegistry) (*APIServer, error) {
	logger, ok := ctx.Value("logger").(*zap.Logger)
	if !ok {
		logger = zap.NewNop()
	}

	server := &APIServer{
		Router:   mux.NewRouter(),
		Logger:   logger,
		Registry: serviceRegistry,
	}

	// Create main API router
	apiRouter := server.Router.PathPrefix("/apis/core/v1").Subrouter()

	// Set up global health check endpoint
	server.Router.HandleFunc("/health", server.healthCheckHandler).Methods("GET")

	// Set up routes for different domains
	if err := server.setupAuthRoutes(apiRouter, ctx); err != nil {
		return nil, err
	}

	if err := server.setupTenantRoutes(apiRouter, ctx); err != nil {
		return nil, err
	}

	if err := server.setupAdminRoutes(apiRouter, ctx); err != nil {
		return nil, err
	}

	logger.Info("API Server initialized with OpenAPI specs")
	return server, nil
}

// setupAuthRoutes creates the auth subrouter using go-apigen
func (s *APIServer) setupAuthRoutes(parent *mux.Router, ctx context.Context) error {
	// Parse OpenAPI spec for auth endpoints
	authSpec, err := generator.ParseOpenAPIFile(filepath.Join(openapiDir, authSpecFile))
	if err != nil {
		s.Logger.Error("Failed to parse auth OpenAPI spec", zap.Error(err))
		return err
	}

	// Get auth handler
	authHandler := authhdlr.NewAuthHandler(s.Registry, s.Logger)

	// Define operations map for auth endpoints
	authOps := generator.OperationMap{
		"loginUser": generator.RouteDefinition{
			Handler: authHandler.Login,
		},
		"registerUser": generator.RouteDefinition{
			Handler: authHandler.Register,
		},
	}

	// Generate router using go-apigen
	authRouter, err := generator.GenerateMuxRouter(authSpec, authOps)
	if err != nil {
		s.Logger.Error("Failed to generate auth router", zap.Error(err))
		return err
	}

	// Mount auth router under /auth path prefix
	parent.PathPrefix("/auth").Handler(
		http.StripPrefix("/apis/core/v1/auth", authRouter),
	)

	s.Logger.Info("Auth routes configured")
	return nil
}

// setupTenantRoutes creates the tenant/onboarding subrouter using go-apigen
func (s *APIServer) setupTenantRoutes(parent *mux.Router, ctx context.Context) error {
	// Parse OpenAPI spec for tenant endpoints
	tenantSpec, err := generator.ParseOpenAPIFile(filepath.Join(openapiDir, tenantSpecFile))
	if err != nil {
		s.Logger.Error("Failed to parse tenant OpenAPI spec", zap.Error(err))
		return err
	}

	// Get onboarding handler
	onboardingHandler := onboardinghdlr.NewOnboardingHandler(s.Registry, s.Logger)

	// Define operations map for tenant endpoints
	tenantOps := generator.OperationMap{
		"onboardTenant": generator.RouteDefinition{
			Handler: onboardingHandler.OnboardTenant,
		},
		"getTenants": generator.RouteDefinition{
			Handler:     onboardingHandler.GetTenants,
			Middlewares: []mux.MiddlewareFunc{middleware.AuthRequiredMiddleware(s.Registry)},
		},
		"getTenantById": generator.RouteDefinition{
			Handler:     onboardingHandler.GetTenantByRequestID,
			Middlewares: []mux.MiddlewareFunc{middleware.AuthRequiredMiddleware(s.Registry)},
		},
		"approveTenant": generator.RouteDefinition{
			Handler:     onboardingHandler.ApproveOnboarding,
			Middlewares: []mux.MiddlewareFunc{middleware.AuthRequiredMiddleware(s.Registry)},
		},
		"checkTenantById": generator.RouteDefinition{
			Handler: onboardingHandler.GetTenantExistsByRequestID,
		},
	}

	// Generate router using go-apigen
	tenantRouter, err := generator.GenerateMuxRouter(tenantSpec, tenantOps)
	if err != nil {
		s.Logger.Error("Failed to generate tenant router", zap.Error(err))
		return err
	}

	// Mount tenant router under /tenants path prefix
	parent.PathPrefix("/tenants").Handler(
		http.StripPrefix("/apis/core/v1/tenants", tenantRouter),
	)

	s.Logger.Info("Tenant routes configured")
	return nil
}

// setupAdminRoutes creates the admin subrouter using go-apigen
func (s *APIServer) setupAdminRoutes(parent *mux.Router, ctx context.Context) error {
	// Parse OpenAPI spec for admin endpoints
	adminSpec, err := generator.ParseOpenAPIFile(filepath.Join(openapiDir, adminSpecFile))
	if err != nil {
		s.Logger.Error("Failed to parse admin OpenAPI spec", zap.Error(err))
		return err
	}

	// Get admin handler
	adminHandler := adminhdlr.NewAdminHandler(s.Registry, s.Logger)

	// Common admin middleware - ensure auth is required for all admin routes
	adminMiddleware := middleware.AuthRequiredMiddleware(s.Registry)

	// Define operations map for admin endpoints
	adminOps := generator.OperationMap{
		"listDepartments": generator.RouteDefinition{
			Handler:     adminHandler.ServeHTTP,
			Middlewares: []mux.MiddlewareFunc{adminMiddleware},
		},
		"createDepartment": generator.RouteDefinition{
			Handler:     adminHandler.ServeHTTP,
			Middlewares: []mux.MiddlewareFunc{adminMiddleware},
		},
		"getDepartmentById": generator.RouteDefinition{
			Handler:     adminHandler.ServeHTTP,
			Middlewares: []mux.MiddlewareFunc{adminMiddleware},
		},
		"updateDepartment": generator.RouteDefinition{
			Handler:     adminHandler.ServeHTTP,
			Middlewares: []mux.MiddlewareFunc{adminMiddleware},
		},
		"deleteDepartment": generator.RouteDefinition{
			Handler:     adminHandler.ServeHTTP,
			Middlewares: []mux.MiddlewareFunc{adminMiddleware},
		},
	}

	// Generate router using go-apigen
	adminRouter, err := generator.GenerateMuxRouter(adminSpec, adminOps)
	if err != nil {
		s.Logger.Error("Failed to generate admin router", zap.Error(err))
		return err
	}

	// Mount admin router under /admin path prefix
	parent.PathPrefix("/admin").Handler(
		http.StripPrefix("/apis/core/v1/admin", adminRouter),
	)

	s.Logger.Info("Admin routes configured")
	return nil
}

// healthCheckHandler provides a simple health check endpoint
func (s *APIServer) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}
