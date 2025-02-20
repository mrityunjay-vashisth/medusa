package handlers

import (
	"log"
	"net/http"

	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
	"go.mongodb.org/mongo-driver/mongo"
)

// MainHandler manages subhandlers dynamically
type MainHandler struct {
	UserHandler       *UserHandler
	OnboardingHandler *OnboardingHandler
	AuthHandler       *AuthHandler
}

// NewMainHandler initializes subhandlers
func NewMainHandler(db *mongo.Client, authClient authpb.AuthServiceClient) *MainHandler {
	return &MainHandler{
		UserHandler:       NewUserHandler(db),
		OnboardingHandler: NewOnboardingHandler(db),
		AuthHandler:       NewAuthHandler(db, authClient),
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
