package adminsvc

import (
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"go.uber.org/zap"
)

type baseService struct {
	db     db.DBClientInterface
	logger *zap.Logger
}

type AdminServicesInterface interface {
	GetCurrentAdmin()
}

type AdminServiceManagerInterface interface {
	Admin() AdminServicesInterface
	Department() DepartmentServicesInterface
}

type adminServiceManager struct {
	adminService      AdminServicesInterface
	departmentService DepartmentServicesInterface
}

type adminService struct {
	*baseService
}

func (m *adminServiceManager) Admin() AdminServicesInterface {
	return m.adminService
}

func (m *adminServiceManager) Department() DepartmentServicesInterface {
	return m.departmentService
}

func NewService(db db.DBClientInterface, registry registry.ServiceRegistry, logger *zap.Logger) AdminServiceManagerInterface {
	base := &baseService{
		db:     db,
		logger: logger,
	}

	return &adminServiceManager{
		adminService: &adminService{baseService: base},
	}
}

func (a *adminService) GetCurrentAdmin() {}
