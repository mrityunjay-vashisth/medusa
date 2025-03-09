package handlers

import (
	"github.com/mrityunjay-vashisth/core-service/internal/services/adminsvc"
	"go.uber.org/zap"
)

type AdminHandlerInterface interface {
}

type adminHandler struct {
	Service adminsvc.AdminServiceManagerInterface
	Logger  *zap.Logger
}

func NewAdminHandler(service adminsvc.AdminServiceManagerInterface, logger *zap.Logger) AdminHandlerInterface {
	return &adminHandler{Service: service, Logger: logger}
}
