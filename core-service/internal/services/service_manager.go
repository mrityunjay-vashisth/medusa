package services

import (
	"context"
	"os"

	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"github.com/mrityunjay-vashisth/core-service/internal/services/adminsvc"
	"github.com/mrityunjay-vashisth/core-service/internal/services/authsvc"
	"github.com/mrityunjay-vashisth/core-service/internal/services/onboardingsvc"
	"go.uber.org/zap"
)

type ServiceManager struct {
	registry registry.ServiceRegistry
	logger   *zap.Logger
}

func NewServiceManager(ctx context.Context, db db.DBClientInterface) *ServiceManager {
	// Create registry
	serviceRegistry := registry.NewServiceRegistry()
	logger, ok := ctx.Value("logger").(*zap.Logger)
	if !ok {
		logger = zap.L()
	}
	authServiceAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authServiceAddr == "" {
		authServiceAddr = "172.26.57.112:50051"
	}
	logger.Info("Initializedddddddddddddddddddd")

	authService := authsvc.NewService(db, authServiceAddr, logger)
	onboardingService := onboardingsvc.NewService(db, serviceRegistry, logger)
	adminService := adminsvc.NewService(db, serviceRegistry, logger)

	serviceRegistry.Register(registry.AuthService, authService)
	serviceRegistry.Register(registry.OnboardingService, onboardingService)
	serviceRegistry.Register(registry.AdminService, adminService)

	return &ServiceManager{
		registry: serviceRegistry,
		logger:   logger,
	}

}

func (sm *ServiceManager) GetRegistry() registry.ServiceRegistry {
	return sm.registry
}

// Helper methods to provide a cleaner API for getting services

// GetAuthService returns the auth service
func (sm *ServiceManager) GetAuthService() authsvc.Service {
	svc, ok := sm.registry.Get(registry.AuthService).(authsvc.Service)
	if !ok {
		// In a real implementation, you might want to handle this error differently
		panic("Auth service not found in registry or has wrong type")
	}
	return svc
}

// GetOnboardingService returns the onboarding service
func (sm *ServiceManager) GetOnboardingService() onboardingsvc.Service {
	svc, ok := sm.registry.Get(registry.OnboardingService).(onboardingsvc.Service)
	if !ok {
		panic("Onboarding service not found in registry or has wrong type")
	}
	return svc
}

// GetAdminService returns the admin service
func (sm *ServiceManager) GetAdminService() adminsvc.Service {
	svc, ok := sm.registry.Get(registry.AdminService).(adminsvc.Service)
	if !ok {
		panic("Admin service not found in registry or has wrong type")
	}
	return svc
}
