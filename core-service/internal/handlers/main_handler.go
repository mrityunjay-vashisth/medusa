package handlers

import (
	"encoding/json"
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
func NewMainHandler(db db.DBClientInterface, newServices *services.Container, logger *zap.Logger) *MainHandler {
	return &MainHandler{
		UserHandler:       NewUserHandler(db),
		OnboardingHandler: NewOnboardingHandler(newServices.OnboardingService, logger),
		AuthHandler:       NewAuthHandler(newServices.AuthService, logger),
	}
}

// GetSubHandler returns the appropriate subhandler
func (h *MainHandler) GetSubHandler(name string) http.HandlerFunc {
	switch name {
	case "UserHandler":
		return h.UserHandler.ServeHTTP
	case "TenantHandler":
		return h.OnboardingHandler.ServeHTTP
	case "AuthHandler":
		return h.AuthHandler.ServeHTTP
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
