package services

import (
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/services/adminsvc"

	"go.uber.org/zap"
)

type ServiceManagerInterface interface {
	GetAuthService() AuthServices
	GetOnboardingService() OnboardingServicesInterface
	GetSessionService() SessionServicesInterface
	GetAdminService() adminsvc.AdminServiceManagerInterface
}

type serviceManager struct {
	authService       *AuthService
	onboardingService OnboardingServicesInterface
	sessionService    SessionServicesInterface
	adminService      adminsvc.AdminServiceManagerInterface
}

func NewServiceManager(db db.DBClientInterface, logger *zap.Logger) ServiceManagerInterface {
	return &serviceManager{
		authService:       NewAuthService("172.26.57.112:50051", logger),
		onboardingService: NewOnboardingService(db, logger),
		sessionService:    NewSessionService(db, logger),
		adminService:      adminsvc.NewAdminServiceManager(db, logger),
	}
}

func (sm *serviceManager) GetAuthService() AuthServices {
	return sm.authService
}

func (sm *serviceManager) GetOnboardingService() OnboardingServicesInterface {
	return sm.onboardingService
}

func (sm *serviceManager) GetSessionService() SessionServicesInterface {
	return sm.sessionService
}

func (sm *serviceManager) GetAdminService() adminsvc.AdminServiceManagerInterface {
	return sm.adminService
}
