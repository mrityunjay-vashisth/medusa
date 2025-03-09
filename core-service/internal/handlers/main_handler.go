package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/services"
	"go.uber.org/zap"
)

type HandlerManagerInterface interface {
	GetSubHandler(name string) http.HandlerFunc
}

// MainHandler manages subhandlers dynamically
type handlerManager struct {
	adminHandler      AdminHandlerInterface
	onboardingHandler OnboardingHandlerInterface
	authHandler       AuthHandlerInterface
}

// NewMainHandler initializes subhandlers
func NewMainHandler(db db.DBClientInterface, newServices services.ServiceManagerInterface, logger *zap.Logger) HandlerManagerInterface {
	return &handlerManager{
		adminHandler:      NewAdminHandler(newServices.GetAdminService(), logger),
		onboardingHandler: NewOnboardingHandler(newServices.GetOnboardingService(), logger),
		authHandler:       NewAuthHandler(newServices.GetAuthService(), logger),
	}
}

func (hm *handlerManager) getAuthHandler() AuthHandlerInterface {
	return hm.authHandler
}

func (hm *handlerManager) getOnboardingHanlder() OnboardingHandlerInterface {
	return hm.onboardingHandler
}

func (hm *handlerManager) getAdminHandler() AdminHandlerInterface {
	return hm.adminHandler
}

// GetSubHandler returns the appropriate subhandler
func (h *handlerManager) GetSubHandler(name string) http.HandlerFunc {
	switch name {
	case "TenantHandler":
		onboardingHandler := h.getOnboardingHanlder()
		return onboardingHandler.ServeHTTP
	case "AuthHandler":
		authHandler := h.getAuthHandler()
		return authHandler.ServeHTTP
	default:
		log.Printf("Unknown handler: %s", name)
		return nil
	}
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
