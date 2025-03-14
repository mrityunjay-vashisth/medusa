package adminhdlr

import (
	"net/http"

	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"go.uber.org/zap"
)

type AdminHandlerInterface interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type adminHandler struct {
	registry registry.ServiceRegistry
	logger   *zap.Logger
}

func NewAdminHandler(registry registry.ServiceRegistry, logger *zap.Logger) AdminHandlerInterface {
	return &adminHandler{
		registry: registry,
		logger:   logger,
	}
}

func (h *adminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
