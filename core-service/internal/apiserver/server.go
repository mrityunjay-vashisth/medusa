package apiserver

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/handlers"
	"github.com/mrityunjay-vashisth/core-service/internal/middleware"
	"github.com/mrityunjay-vashisth/core-service/internal/services"
	"go.uber.org/zap"
)

// List of routes that should bypass authentication
var publicRoutes = []string{
	"/apis/core/v1/auth/login",
	"/apis/core/v1/auth/register",
	"/openapi/v2",
}

// APIServer initializes API routes dynamically
type APIServer struct {
	Router *mux.Router
	Logger *zap.Logger
}

// NewAPIServer loads `registry.json` and registers API routes
func NewAPIServer(db db.DBClientInterface, registeredServices *services.ServiceTypes) *APIServer {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	server := &APIServer{
		Router: mux.NewRouter(),
		Logger: logger,
	}
	// Load API registry
	registry, err := LoadRegistry()
	if err != nil {
		log.Fatalf("Failed to load API registry: %v", err)
	}

	// Initialize handlers
	mainHandler := handlers.NewMainHandler(db, registeredServices, logger)

	// Loop through `registry.json` and create API groups dynamically
	for group, versions := range registry {
		for version, resources := range versions {
			apiPath := "/apis/" + group + "/" + version
			subRouter := server.Router.PathPrefix(apiPath).Subrouter()

			for path, resource := range resources {
				fullPath := apiPath + path
				handlerFunc := mainHandler.GetSubHandler(resource.Subhandler)

				if handlerFunc == nil {
					log.Printf("Error: Handler %s not found for %s", resource.Subhandler, fullPath)
					continue
				}

				subRouter.HandleFunc(path, handlerFunc).Methods(resource.Methods...)
				log.Printf("Registered API: %s [%v]", fullPath, resource.Methods)
			}
		}
	}
	// authClient := newServices.AuthService.GetClient()
	// Apply middleware
	server.Router.Use(middleware.LoggingMiddleware(logger))
	server.Router.Use(middleware.ConditionalAuthMiddleware(registeredServices, publicRoutes))
	log.Println("API Server started on /apis/core/v1")
	return server
}
