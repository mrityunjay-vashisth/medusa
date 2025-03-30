package authhdlr

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mrityunjay-vashisth/core-service/internal/handlers/utility"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"github.com/mrityunjay-vashisth/core-service/internal/services/authsvc"
	"go.uber.org/zap"
)

type AuthHandlerInterface interface {
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
		TenantID string `json:"tenantid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		// http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authService, err := a.getAuthService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	authResp, err := authService.Login(r.Context(), req.Username, req.Password, req.TenantID)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
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

// Register handles user registration requests
func (a *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.AuthRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" || req.Email == "" || req.Role == "" {
		utility.RespondWithError(w, http.StatusBadRequest, "Missing required fields: username, password, email, and role are required")
		return
	}

	// Get auth service
	authService, err := a.getAuthService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	authResp, err := authService.Register(r.Context(), req)

	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	utility.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"message": authResp.Message,
	})
}
