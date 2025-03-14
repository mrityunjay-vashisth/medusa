package authhdlr

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/handlers/utility"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"github.com/mrityunjay-vashisth/core-service/internal/services/authsvc"
	"go.uber.org/zap"
)

type AuthHandlerInterface interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
}

type authHandler struct {
	registry registry.ServiceRegistry
	logger   *zap.Logger
}

func NewAuthHandler(registry registry.ServiceRegistry, logger *zap.Logger) AuthHandlerInterface {
	return &authHandler{
		registry: registry,
		logger:   logger,
	}
}

// ServeHTTP routes requests
func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.logger.Info("Auth API", zap.String("subpath", vars["subpath"]))
	subPath := vars["subpath"]

	switch subPath {
	case "login":
		h.Login(w, r)
	case "register":
		h.Register(w, r)
	default:
		utility.RespondWithError(w, http.StatusNotFound, "Invalid Onboarding API")
	}
}

// getAuthService retrieves the auth service from the registry
func (h *authHandler) getAuthService() (authsvc.Service, error) {
	service, ok := h.registry.Get(registry.AuthService).(authsvc.Service)
	if !ok {
		h.logger.Error("Failed to get auth service from registry")
		return nil, errors.New("internal service error")
	}
	return service, nil
}

func (a *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		// http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authService, err := a.getAuthService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}

	authResp, err := authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if authResp.Token == "" {
		utility.RespondWithError(w, http.StatusUnauthorized, authResp.Message)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"token":   authResp.Token,
		"message": authResp.Message,
		"email":   authResp.Email,
	})
}

func (a *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	// handle reg request
}
