[![Go Test](https://github.com/mrityunjay-vashisth/medusa/actions/workflows/go-test.yml/badge.svg)](https://github.com/mrityunjay-vashisth/medusa/actions/workflows/go-test.yml)

# Medusa Project Developer Guide

Welcome to the Medusa project! This guide will help you understand the project structure, set up your development environment, and guide you through contributing to the codebase.

## Table of Contents

**1. Architecture Overview**
**2. Environment Setup**
**3. Project Structure**
**4. Development Workflow**
**5. Adding New Features**
**6. Testing Guidelines**
**7. Common Patterns**
**8. Troubleshooting**

## 1. Architecture Overview

Medusa is a microservices-based system with the following main components:

- **auth-service**: Handles authentication, authorization, and user management
- **core-service**: Contains the main business logic, API endpoints, and data management

The services communicate via gRPC and share a MongoDB database. The application follows a clean architecture pattern with separated concerns:

```
Client Request → API Routes → Handlers → Services → Database
```

## 2. Environment Setup

### Prerequisites
- Go 1.23+
- Docker and Docker Compose
- MongoDB (or use the containerized version)
- Make

### Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/mrityunjay-vashisth/medusa.git
   cd medusa
   ```

2. Start the services (automatically handles MongoDB):
   ```bash
   make up-auto
   ```

3. For development, you can run individual services:
   ```bash
   # For auth service
   cd auth-service
   go run cmd/main.go

   # For core service
   cd core-service
   go run cmd/main.go
   ```

### Environment Variables

Key environment variables are located in `.env` files in each service directory. For local development, the default values should work, but you can modify them as needed.

## 3. Project Structure

### Common Directory Structure

Each service follows a similar structure:

```
service-name/
├── cmd/                  # Application entry points
├── internal/             # Private application code
│   ├── middleware/       # HTTP middleware
│   ├── handlers/         # HTTP handlers
│   ├── services/         # Business logic
│   ├── models/           # Data models and DTOs
│   ├── db/               # Database access
│   └── config/           # Configuration
├── test/                 # Test utilities and mocks
├── Dockerfile            # Container definition
└── go.mod, go.sum        # Go modules
```

### Key Components

- **Registry Pattern**: Services are registered and accessed through a central registry (see `core-service/internal/registry/registry.go`)
- **Interface-Driven Design**: Components interact through interfaces for loose coupling
- **Middleware Chain**: HTTP requests flow through middleware for logging, auth, recovery, etc.
- **API Registry**: API endpoints are defined in `core-service/internal/apiserver/registry.json`

## 4. Development Workflow

### Adding New API Endpoints

1. Define the endpoint in `core-service/internal/apiserver/registry.json`:
   ```json
   "/my-resource/{action}": {
     "methods": ["GET", "POST"],
     "subhandler": "MyResourceHandler"
   }
   ```

2. Create a handler in `core-service/internal/handlers/my_resource_handler.go`:
   ```go
   type MyResourceHandlerInterface interface {
       ServeHTTP(w http.ResponseWriter, r *http.Request)
       // Add other methods needed
   }

   type myResourceHandler struct {
       registry registry.ServiceRegistry
       logger   *zap.Logger
   }

   func NewMyResourceHandler(registry registry.ServiceRegistry, logger *zap.Logger) MyResourceHandlerInterface {
       return &myResourceHandler{
           registry: registry,
           logger:   logger,
       }
   }

   func (h *myResourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
       // Handle routing based on path variables
   }
   ```

3. Update the main handler in `core-service/internal/handlers/main_handler.go` to include your handler:
   ```go
   func (h *handlerManager) GetSubHandler(name string) http.HandlerFunc {
       switch name {
       // ...existing cases
       case "MyResourceHandler":
           return NewMyResourceHandler(h.registry, h.logger).ServeHTTP
       // ...
       }
   }
   ```

4. Create a service in `core-service/internal/services/myresourcesvc/my_resource_service.go`
5. Register the service in `core-service/internal/services/service_manager.go`

### Adding New Models

1. Create the model in `core-service/internal/models/my_model.go`:
   ```go
   package models

   import "time"

   type MyModel struct {
       ID          string    `bson:"_id"`
       Name        string    `bson:"name"`
       CreatedAt   time.Time `bson:"created_at"`
       // Other fields
   }
   ```

2. Use the model in your service and handlers

## 5. Adding New Features

### Service Layer

1. Define the service interface in a new file under `internal/services/myservicesvc/my_service.go`:
   ```go
   package myservicesvc

   import (
       "context"
       "github.com/mrityunjay-vashisth/core-service/internal/db"
       "github.com/mrityunjay-vashisth/core-service/internal/models"
       "github.com/mrityunjay-vashisth/core-service/internal/registry"
       "go.uber.org/zap"
   )

   type Service interface {
       DoSomething(ctx context.Context, param string) (interface{}, error)
   }

   type myService struct {
       db     db.DBClientInterface
       logger *zap.Logger
   }

   func NewService(db db.DBClientInterface, registry registry.ServiceRegistry, logger *zap.Logger) Service {
       return &myService{
           db:     db,
           logger: logger,
       }
   }

   func (s *myService) DoSomething(ctx context.Context, param string) (interface{}, error) {
       // Implement business logic
   }
   ```

2. Register your service in `service_manager.go`:
   ```go
   // Add constant in registry/service_names.go
   const MyService = "my_service"

   // Register in service_manager.go
   myService := myservicesvc.NewService(db, serviceRegistry, logger)
   serviceRegistry.Register(registry.MyService, myService)
   ```

### Database Operations

When interacting with the database:

1. Use the `db.DBClientInterface` methods:
   ```go
   // Create a document
   data := map[string]interface{}{
       "name": "Example",
       "created_at": time.Now(),
   }
   result, err := s.db.Create(
       ctx, 
       data,
       db.WithDatabaseName("coredb"),
       db.WithCollectionName("my_collection"),
   )

   // Read a document
   filter := bson.M{"name": "Example"}
   result, err := s.db.Read(
       ctx, 
       filter,
       db.WithDatabaseName("coredb"),
       db.WithCollectionName("my_collection"),
   )
   ```

2. Always use the options pattern with `WithDatabaseName` and `WithCollectionName`
3. Handle errors appropriately and log them using the provided logger

## 6. Testing Guidelines

### Unit Testing

1. For each component, create a test file in the same package with `_test.go` suffix
2. Use the provided mocks in `test/mocks/` directory
3. Example test:
   ```go
   func TestMyService_DoSomething(t *testing.T) {
       // Setup
       mockDB := &mocks.MockDBClient{
           ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
               return map[string]interface{}{"name": "Test"}, nil
           },
       }
       logger, _ := zap.NewDevelopment()
       service := NewService(mockDB, nil, logger)

       // Execute
       result, err := service.DoSomething(context.Background(), "test")

       // Assert
       assert.NoError(t, err)
       assert.NotNil(t, result)
   }
   ```

### MongoDB Testing

For MongoDB integration tests, use the testing package from MongoDB:

```go
func TestMongoIntegration(t *testing.T) {
    mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

    mt.Run("test name", func(mt *mtest.T) {
        // Mock responses
        mt.AddMockResponses(mtest.CreateSuccessResponse())
        
        // Setup 
        client := db.NewDBClient(db.DBConfig{
            Type: db.MongoDB,
            URI: "mongodb://fake-uri",
        })
        client.mongoClient.client = mt.Client
        
        // Test logic
    })
}
```

## 7. Common Patterns

### Error Handling

- Always return descriptive errors: `return nil, errors.New("specific error message")`
- Log errors at the appropriate level:
  ```go
  if err != nil {
      logger.Error("Failed to fetch data", zap.Error(err), zap.String("param", value))
      return nil, errors.New("failed to process request")
  }
  ```
- Use the common response utilities:
  ```go
  if err != nil {
      common.RespondWithError(w, r, logger, err, http.StatusInternalServerError, "Failed to process request")
      return
  }
  common.RespondWithJSON(w, r, logger, data, http.StatusOK, "Success")
  ```

### Service Registry Usage

- Get services from the registry:
  ```go
  service, ok := h.registry.Get(registry.MyService).(myservicesvc.Service)
  if !ok {
      logger.Error("Failed to get service from registry")
      return nil, errors.New("internal service error")
  }
  ```

### Authentication

- For authenticated endpoints, extract the token:
  ```go
  token := r.Header.Get("Authorization")
  claims, err := validateToken(token)
  if err != nil {
      respondWithError(w, http.StatusUnauthorized, "Invalid token")
      return
  }
  ```

## 8. Troubleshooting

### Common Issues

1. **Service Connection Issues**
   - Check if MongoDB is running: `nc -z localhost 27017`
   - Verify auth service is running: `nc -z localhost 50051`
   - Ensure environment variables are correct

2. **Build Errors**
   - Run `go mod tidy` to fix dependency issues
   - Check Go version is 1.23 or higher

3. **Authorization Errors**
   - Ensure JWT_SECRET_KEY is consistent between services
   - Check token expiration times

### Debugging

- Use the structured logging:
  ```go
  logger.Debug("Processing request", 
      zap.String("request_id", requestID),
      zap.Any("data", data))
  ```

- For Docker environments, check logs:
  ```bash
  make logs-auth  # View auth service logs
  make logs-core  # View core service logs
  ```

## Getting Help

If you're stuck or have questions, please:
1. Check existing documentation and code comments
2. Review similar implementations in the codebase
3. Reach out to the team for assistance

Happy coding!