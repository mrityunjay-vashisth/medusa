package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"go.uber.org/zap"
)

type HandlerManagerInterface interface {
	GetSubHandler(name string) http.HandlerFunc
}

// MainHandler manages subhandlers dynamically
type handlerManager struct {
	db       db.DBClientInterface
	registry registry.ServiceRegistry
	logger   *zap.Logger
}

// NewMainHandler initializes subhandlers
func NewMainHandler(db db.DBClientInterface, registry registry.ServiceRegistry, logger *zap.Logger) HandlerManagerInterface {
	return &handlerManager{
		db:       db,
		registry: registry,
		logger:   logger,
	}
}

// GetSubHandler returns the appropriate subhandler
func (h *handlerManager) GetSubHandler(name string) http.HandlerFunc {
	switch name {
	case "TenantHandler":
		return NewOnboardingHandler(h.registry, h.logger).ServeHTTP
	case "AuthHandler":
		return NewAuthHandler(h.registry, h.logger).ServeHTTP
	case "AdminHandler":
		return NewAdminHandler(h.registry, h.logger).ServeHTTP

	default:
		h.logger.Error("Unknown handler", zap.String("name", name))
		return nil
	}
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
