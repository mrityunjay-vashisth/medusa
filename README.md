# Medusa Project Developer Guide

Welcome to the Medusa project! This guide will help you understand the project structure, set up your development environment, and guide you through contributing to the codebase.

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Environment Setup](#2-environment-setup)
3. [Project Structure](#3-project-structure)
4. [Development Workflow](#4-development-workflow)
5. [Adding New Features](#5-adding-new-features)
6. [Testing Guidelines](#6-testing-guidelines)
7. [Common Patterns](#7-common-patterns)
8. [Troubleshooting](#8-troubleshooting)

## 1. Architecture Overview

Medusa is a microservices-based healthcare management system with the following main components:

- **auth-service**: Handles authentication, authorization, and user management via gRPC
- **core-service**: Contains the main business logic, API endpoints, and data management

The services communicate via gRPC and share MongoDB databases. The application follows a clean architecture pattern with separated concerns:

```
Client Request → API Routes → Handlers → Services → Database
```

### Key Design Decisions

- **OpenAPI-Driven Development**: API specifications are defined in OpenAPI YAML files
- **Service Registry Pattern**: Components are registered and accessed through centralized registries
- **Interface-Based Design**: All components interact through interfaces for easier testing and flexibility
- **MongoDB for Persistence**: Using MongoDB for flexible data storage without rigid schemas

## 2. Environment Setup

### Prerequisites

- Go 1.23+ (currently using 1.24.0)
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

Key environment variables are located in `.env` files in each service directory:

**Auth Service**:
```
MONGO_URI=mongodb://mongodb:27017
MONGO_DB_NAME=authdb
JWT_SECRET_KEY=your-secure-jwt-secret-replace-in-production
JWT_EXPIRATION_HOURS=24
AUTH_GRPC_PORT=50051
```

**Core Service**:
```
MONGO_URI=mongodb://mongodb:27017
MONGO_DB_NAME=coredb
API_PORT=8080
AUTH_SERVICE_ADDR=auth-service:50051
JWT_SECRET_KEY=your-secure-jwt-secret-replace-in-production
```

Important: The JWT secret key must match between auth and core services.

## 3. Project Structure

### Common Directory Structure

Each service follows a similar structure:

```
service-name/
├── cmd/                  # Application entry points
├── internal/             # Private application code
│   ├── middleware/       # HTTP middleware
│   ├── handlers/         # HTTP/gRPC handlers
│   ├── services/         # Business logic
│   ├── models/           # Data models and DTOs
│   ├── db/               # Database access
│   ├── config/           # Configuration
│   │   └── openapi/      # OpenAPI specifications
├── test/                 # Test utilities and mocks
├── Dockerfile            # Container definition
└── go.mod, go.sum        # Go modules
```

### Key Components

- **Registry Pattern**: Services are registered and accessed through a central registry (see `core-service/internal/registry/registry.go`)
- **Interface-Driven Design**: Components interact through interfaces for loose coupling
- **Middleware Chain**: HTTP requests flow through middleware for logging, auth, recovery, etc.
- **OpenAPI Specifications**: API endpoints are defined in `core-service/internal/config/openapi/*.yaml`

### Important Packages

- **db**: Database abstraction layer with MongoDB implementation
- **models**: Core domain models like User, Patient, Doctor, Admin
- **services**: Business logic implementation
- **handlers**: API endpoint handlers
- **middleware**: HTTP request processing middleware
- **registry**: Service discovery and registration

## 4. Development Workflow

### Working with OpenAPI Specifications

The project uses go-apigen to generate router configurations from OpenAPI specs. The specs are located in `core-service/internal/config/openapi/`.

To define a new API endpoint:

1. Add it to the appropriate OpenAPI specification file
2. The router generation is handled in `core-service/internal/apiserver/server.go`

### Handler Implementation

Each handler should:
1. Get dependencies from the service registry
2. Process the request
3. Call the appropriate service method
4. Return the formatted response

Example:

```go
func (h *myHandler) HandleEndpoint(w http.ResponseWriter, r *http.Request) {
    // Get service from registry
    service, err := h.getMyService()
    if err != nil {
        utility.RespondWithError(w, http.StatusInternalServerError, "Internal service error")
        return
    }
    
    // Process request
    var req MyRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utility.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    // Call service
    result, err := service.DoSomething(r.Context(), req)
    if err != nil {
        utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    // Return response
    utility.RespondWithJSON(w, http.StatusOK, result)
}
```

## 5. Adding New Features

### Adding a New Domain Model

1. Create a model in `core-service/internal/models/my_model.go`:
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

### Adding a New Service

1. Define the service interface in `internal/services/myservicesvc/my_service.go`:
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

### Adding a New API Endpoint

1. Define the endpoint in an OpenAPI spec file (e.g., `core-service/internal/config/openapi/my_domain.yaml`)

2. Create a handler in `core-service/internal/handlers/mydomainhdlr/my_domain_handler.go`:
   ```go
   package mydomainhdlr

   import (
       "net/http"
       "github.com/mrityunjay-vashisth/core-service/internal/registry"
       "go.uber.org/zap"
   )

   type MyDomainHandlerInterface interface {
       HandleEndpoint(w http.ResponseWriter, r *http.Request)
   }

   type myDomainHandler struct {
       registry registry.ServiceRegistry
       logger   *zap.Logger
   }

   func NewMyDomainHandler(registry registry.ServiceRegistry, logger *zap.Logger) MyDomainHandlerInterface {
       return &myDomainHandler{
           registry: registry,
           logger:   logger,
       }
   }

   func (h *myDomainHandler) HandleEndpoint(w http.ResponseWriter, r *http.Request) {
       // Implementation
   }
   ```

3. Register the handler in the API server setup (`apiserver/server.go`)

### Working with the Database

The database abstraction layer provides a consistent interface for all operations:

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

// Read multiple documents
filter := bson.M{"status": "active"}
results, err := s.db.ReadAll(
    ctx, 
    filter,
    db.WithDatabaseName("coredb"),
    db.WithCollectionName("my_collection"),
)

// Delete documents
filter := bson.M{"expired": true}
count, err := s.db.Delete(
    ctx, 
    filter,
    db.WithDatabaseName("coredb"),
    db.WithCollectionName("my_collection"),
)
```

Always use the options pattern with `WithDatabaseName` and `WithCollectionName` for clarity.

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

### Test Coverage

The project uses GitHub Actions to track test coverage. Run tests locally with:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # View coverage in browser
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
      utility.RespondWithError(w, r, logger, err, http.StatusInternalServerError, "Failed to process request")
      return
  }
  utility.RespondWithJSON(w, r, logger, data, http.StatusOK, "Success")
  ```

### Service Registry Usage

```go
service, ok := h.registry.Get(registry.MyService).(myservicesvc.Service)
if !ok {
    logger.Error("Failed to get service from registry")
    return nil, errors.New("internal service error")
}
```

### Authentication

For authenticated endpoints, extract and validate the token:
```go
// In handler
token := r.Header.Get("Authorization")
claims, err := validateToken(token)
if err != nil {
    utility.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
    return
}

// Check role if needed
if claims.Role != "admin" {
    utility.RespondWithError(w, http.StatusForbidden, "Insufficient permissions")
    return
}
```

Or use the authentication middleware:
```go
// In router setup
router.Handle("/secure-endpoint", middleware.AuthRequiredMiddleware(registry)(handler))
```

### Validation

For input validation, use appropriate checks before processing:
```go
if req.OrganizationName == "" || req.Email == "" {
    return "", errors.New("organization name and email are required")
}
```

For more complex APIs, consider using the validation middleware in `middleware/validation_middleware.go`.

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

- Check MongoDB data directly:
  ```bash
  mongo mongodb://localhost:27017
  > use coredb
  > db.onboarding_requests.find()
  ```

### Common Error Messages & Solutions

- **"unsupported database type"**: Check MongoDB connection string and ensure MongoDB is running
- **"failed to get service from registry"**: Ensure service is properly registered in service_manager.go
- **"invalid token"**: Check JWT secret consistency and token expiration
- **"onboarding request already exists"**: Email is already registered for onboarding

## Getting Help

If you're stuck or have questions, please:
1. Check existing documentation and code comments
2. Review similar implementations in the codebase
3. Reach out to the team for assistance

Happy coding!