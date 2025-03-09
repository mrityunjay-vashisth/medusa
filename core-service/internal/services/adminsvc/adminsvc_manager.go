package adminsvc

import (
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"go.uber.org/zap"
)

type Service interface {
	GetCurrentAdmin()
}

type adminService struct {
	db          db.DBClientInterface
	logger      *zap.Logger
	svcRegistry registry.ServiceRegistry
}

func NewService(db db.DBClientInterface, registry registry.ServiceRegistry, logger *zap.Logger) Service {
	return &adminService{
		db:          db,
		svcRegistry: registry,
		logger:      logger,
	}
}

func (a *adminService) GetCurrentAdmin() {}
