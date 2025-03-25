package onboardinghdlr

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/handlers/utility"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"github.com/mrityunjay-vashisth/core-service/internal/services/authsvc"
	"github.com/mrityunjay-vashisth/core-service/internal/services/onboardingsvc"
	"go.uber.org/zap"
)

var jwtKey = []byte("your-secure-jwt-secret-replace-in-production")

type OnboardingHandlerInterface interface {
	OnboardTenant(w http.ResponseWriter, r *http.Request)
	GetTenants(w http.ResponseWriter, r *http.Request)
	GetTenantByRequestID(w http.ResponseWriter, r *http.Request)
	ApproveOnboarding(w http.ResponseWriter, r *http.Request)
}

// OnboardingHandler handles all onboarding-related requests
type onboardingHandler struct {
	registry registry.ServiceRegistry
	logger   *zap.Logger
}

func NewOnboardingHandler(registry registry.ServiceRegistry, logger *zap.Logger) OnboardingHandlerInterface {
	return &onboardingHandler{
		registry: registry,
		logger:   logger,
	}
}

// getOnboardingService retrieves the onboarding service from the registry
func (h *onboardingHandler) getOnboardingService() (onboardingsvc.Service, error) {
	service, ok := h.registry.Get(registry.OnboardingService).(onboardingsvc.Service)
	if !ok {
		h.logger.Error("Failed to get onboarding service from registry")
		return nil, errors.New("internal service error")
	}
	return service, nil
}

// getAuthService retrieves the auth service from the registry
func (h *onboardingHandler) getAuthService() (authsvc.Service, error) {
	service, ok := h.registry.Get(registry.AuthService).(authsvc.Service)
	if !ok {
		h.logger.Error("Failed to get auth service from registry")
		return nil, errors.New("internal service error")
	}
	return service, nil
}

// OnboardTenant handles onboarding requests
func (h *onboardingHandler) OnboardTenant(w http.ResponseWriter, r *http.Request) {
	var req models.OnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.OrganizationName == "" || req.Email == "" || req.Role == "" {
		utility.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.logger.Info("Received OnboardTenant request",
		zap.String("Organization Name", req.OrganizationName),
		zap.String("Email", req.Email),
		zap.String("role", req.Role),
	)

	service, err := h.getOnboardingService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}

	requestId, err := service.OnboardTenant(r.Context(), req)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("Onboarding request submitted", zap.String("request_id", requestId))
	respData := map[string]string{
		"message":    "Onboarding request submitted, pending approval",
		"request_id": requestId,
	}
	utility.RespondWithJSON(w, http.StatusOK, respData)
}

// GetPendingRequests fetches pending onboarding requests
func (h *onboardingHandler) GetTenants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utility.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}
	token := r.Header.Get("Authorization")
	if token == "" {
		utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if !validateServiceToken(w, token) {
		utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	service, err := h.getOnboardingService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}

	status := r.URL.Query().Get("state")
	requests, err := service.GetTenants(r.Context(), status)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if requests == nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "No pending request for approval")
		return
	}
	utility.RespondWithJSON(w, http.StatusOK, requests)
}

func (h *onboardingHandler) GetTenantByRequestID(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		h.logger.Info("Empty token received")
		return
	}
	if !validateServiceToken(w, token) {
		h.logger.Info("Token validation failed")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	// Validate the ID parameter
	if id == "" {
		utility.RespondWithError(w, http.StatusBadRequest, "Missing tenant ID")
		return
	}

	service, err := h.getOnboardingService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}

	requests, err := service.GetTenantByID(r.Context(), id)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	json.NewEncoder(w).Encode(requests)
}

// ApproveOnboarding approves onboarding requests
func (h *onboardingHandler) ApproveOnboarding(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if !validateServiceToken(w, token) {
		utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.RequestID == "" {
		utility.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	service, err := h.getOnboardingService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}

	// Get tenant data first to extract registration info
	tenantData, err := service.GetTenantByID(r.Context(), req.RequestID)
	if err != nil {
		h.logger.Info("Failed to fetch tenant data", zap.Error(err), zap.String("request_id", req.RequestID))
		utility.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch tenant data: "+err.Error())
		return
	}

	// Convert to map[string]interface{} if it isn't already
	tenantMap, ok := tenantData.(map[string]interface{})
	if !ok {
		h.logger.Info("Invalid tenant data format", zap.Any("tenant_data", tenantData))
		utility.RespondWithError(w, http.StatusInternalServerError, "Invalid tenant data format")
		return
	}

	// Extract required fields for registration
	email, _ := tenantMap["email"].(string)
	username, _ := tenantMap["username"].(string)
	role, _ := tenantMap["role"].(string)
	organizationName, _ := tenantMap["organization_name"].(string)
	tenantID, _ := tenantMap["tenant_id"].(string)

	if email == "" || username == "" || role == "" || tenantID == "" {
		h.logger.Info("Missing required tenant data for registration",
			zap.String("email", email),
			zap.String("username", username),
			zap.String("role", role),
			zap.String("tenantID", tenantID))
		utility.RespondWithError(w, http.StatusInternalServerError, "Tenant data missing required fields")
		return
	}

	err = service.ApproveOnboarding(r.Context(), req.RequestID)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get auth service to register the user
	authService, err := h.getAuthService()
	if err != nil {
		h.logger.Error("Failed to get auth service", zap.Error(err))
		utility.RespondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}

	// Create a registration request
	regRequest := models.AuthRegisterRequest{
		Username: username,
		// Password is generated in ApproveOnboarding method and accessible via logs
		// For production, implement a better password delivery mechanism
		Email:    email,
		Name:     organizationName, // Using org name as user name initially
		Role:     role,
		TenantId: tenantID,
	}

	// Register the user with auth service
	_, err = authService.Register(r.Context(), regRequest)
	if err != nil {
		h.logger.Error("Failed to register user with auth service",
			zap.Error(err),
			zap.String("username", username),
			zap.String("email", email))
		// Continue despite error since onboarding is already approved
		// In production, you might want to handle this differently
		utility.RespondWithError(w, http.StatusInternalServerError, "Failed to create account, contact support")
		return
	}

	// Return success response
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Onboarding request approved and user registered",
		"email":     email,
		"username":  username,
		"tenant_id": tenantID,
	})
}

func validateServiceToken(w http.ResponseWriter, tokenString string) bool {
	ownerClaims, err := validateToken(tokenString)
	if err != nil {
		log.Println(err.Error())
		utility.RespondWithError(w, http.StatusUnauthorized, "Invalid owner token")
		return false
	}
	log.Printf("%s", ownerClaims.Role)
	if ownerClaims.Role != "superuser" {
		utility.RespondWithError(w, http.StatusForbidden, "Unauthorized: Only superusers can view/approve")
		return false
	}
	return true
}

func validateToken(tokenString string) (*models.UserClaims, error) {
	claims := &models.UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		log.Println(err.Error())
		return nil, errors.New("invalid token")
	}
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}
