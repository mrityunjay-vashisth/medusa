package handlers

import (
	"log"
	"net/http"

	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/services"
	"go.uber.org/zap"
)

// MainHandler manages subhandlers dynamically
type MainHandler struct {
	UserHandler       *UserHandler
	OnboardingHandler *OnboardingHandler
	AuthHandler       *AuthHandler
}

// NewMainHandler initializes subhandlers
func NewMainHandler(db db.DBClientInterface, newServices *services.ServiceTypes, logger *zap.Logger) *MainHandler {
	return &MainHandler{
		UserHandler:       NewUserHandler(db),
		OnboardingHandler: NewOnboardingHandler(newServices.OnboardingService, logger),
		AuthHandler:       NewAuthHandler(db, newServices.AuthService.GetClient()),
	}
}

// GetSubHandler returns the appropriate subhandler
func (h *MainHandler) GetSubHandler(name string) http.HandlerFunc {
	switch name {
	case "UserHandler":
		return h.UserHandler.ServeHTTP
	case "OnboardingHandler":
		return h.OnboardingHandler.ServeHTTP
	case "AuthHandler":
		return h.AuthHandler.ServeHTTP
	default:
		log.Printf("Unknown handler: %s", name)
		return nil
	}
}
