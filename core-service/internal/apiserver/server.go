package apiserver

import (
	"context"
	"log"

	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/config"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/handlers"
	"github.com/mrityunjay-vashisth/core-service/internal/middleware"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"go.uber.org/zap"
)

// List of routes that should bypass authentication
var publicRoutes = []string{
	"/apis/core/v1/auth/login",
	"/apis/core/v1/auth/register",
	"/apis/core/v1/tenants",
	"/openapi/v2",
}

// APIServer initializes API routes dynamically
type APIServer struct {
	Router   *mux.Router
	Logger   *zap.Logger
	Registry registry.ServiceRegistry
}

// NewAPIServer loads `registry.json` and registers API routes
func NewAPIServer(ctx context.Context, db db.DBClientInterface, serviceRegistry registry.ServiceRegistry) *APIServer {
	logger, ok := ctx.Value("logger").(*zap.Logger)
	if !ok {
		// Fallback if logger is not in context or is of wrong type
		logger = zap.NewNop() // or some default logger
	}

	server := &APIServer{
		Router:   mux.NewRouter(),
		Logger:   logger,
		Registry: serviceRegistry,
	}

	// Initialize handlers
	mainHandler := handlers.NewMainHandler(ctx, db, serviceRegistry)

	// Loop through `registry.json` and create API groups dynamically
	for group, versions := range config.Registry {
		for version, resources := range versions {
			apiPath := "/apis/" + group + "/" + version
			subRouter := server.Router.PathPrefix(apiPath).Subrouter()
			logger.Info("Handler details",
				zap.String("apiPath", apiPath))

			for path, resource := range resources {
				fullPath := apiPath + path
				handlerFunc := mainHandler.GetSubHandler(resource.Subhandler)
				logger.Info("Handler found",
					zap.String("handler", resource.Subhandler),
					zap.String("fullpath", fullPath),
					zap.String("apiPath", apiPath),
					zap.String("path", path))

				if handlerFunc == nil {
					logger.Error("Handler not found",
						zap.String("handler", resource.Subhandler),
						zap.String("path", fullPath))
					continue
				}

				subRouter.HandleFunc(path, handlerFunc).Methods(resource.Methods...)
				log.Printf("Registered API: %s [%v]", fullPath, resource.Methods)
			}
		}
	}
	// Apply middleware
	server.Router.Use(middleware.RecoveryMiddleware(logger))
	server.Router.Use(middleware.LoggingMiddleware(logger))
	server.Router.Use(middleware.ConditionalAuthMiddleware(serviceRegistry, config.PublicRoutes))
	log.Println("API Server started on /apis/core/v1")
	return server
}
