package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"github.com/mrityunjay-vashisth/core-service/internal/services"
	"go.uber.org/zap"
)

var jwtKey = []byte("my-secret-key")

type OnboardingHandlerInterface interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	OnboardTenant(w http.ResponseWriter, r *http.Request)
	GetTenants(w http.ResponseWriter, r *http.Request)
	GetTenantByRequestID(w http.ResponseWriter, r *http.Request, id string)
	ApproveOnboarding(w http.ResponseWriter, r *http.Request)
}

// OnboardingHandler handles all onboarding-related requests
type onboardingHandler struct {
	Service services.OnboardingServicesInterface
	Logger  *zap.Logger
}

func NewOnboardingHandler(service services.OnboardingServicesInterface, logger *zap.Logger) OnboardingHandlerInterface {
	return &onboardingHandler{Service: service, Logger: logger}
}

/*
POST 	/tenants/{action}
GET 	/tenants/{action}
GET 	/tenants/{action}/{id}
PATCH 	/tenants/{action}/{id}

POST	/tenants/onboard
GET    	/tenants/status?state={pending, active}
GET    	/tenants/viewall
GET    	/tenants/view/{id}?search=reqid
GET    	/tenants/view/{id}?search=tenantid
GET    	/tenants/view/{id}?search=reqid
GET    	/tenants/view/{id}?search=reqid
PATCH  	/tenants/approve/{id}
*/

// ServeHTTP routes requests to onboarding-specific functions
func (h *onboardingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.Logger.Info("Onboarding API", zap.String("action", vars["action"]))
	h.Logger.Info("Onboarding API", zap.String("id", vars["id"]))
	action := vars["action"]
	id := vars["id"]

	switch r.Method {
	case http.MethodPost:
		switch action {
		case "onboard":
			h.OnboardTenant(w, r)
		default:
			respondWithError(w, http.StatusNotFound, "Invalid Onboarding POST API")
		}
	case http.MethodGet:
		switch action {
		case "status":
			if id == "" {
				h.GetTenants(w, r)
			} else {
				h.GetTenantByRequestID(w, r, id)
			}
		default:
			respondWithError(w, http.StatusNotFound, "Invalid Onboarding GET API")
		}
	case http.MethodPatch:
		switch action {
		case "approve":
			h.ApproveOnboarding(w, r)
		default:
			respondWithError(w, http.StatusNotFound, "Invalid Onboarding PATCH API")
		}
	default:
		respondWithError(w, http.StatusNotFound, "Invalid Onboarding API Method")

	}
}

// OnboardTenant handles onboarding requests
func (h *onboardingHandler) OnboardTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}
	var req models.OnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	requestId, err := h.Service.OnboardTenant(r.Context(), req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.Logger.Info("Onboarding request submitted", zap.String("request_id", requestId))
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Onboarding request submitted, pending approval",
		"request_id": requestId,
	})
}

// GetPendingRequests fetches pending onboarding requests
func (h *onboardingHandler) GetTenants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}
	token := r.Header.Get("Authorization")
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if !validateServiceToken(w, token) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	status := r.URL.Query().Get("state")
	requests, err := h.Service.GetTenants(r.Context(), status)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	json.NewEncoder(w).Encode(requests)
}

func (h *onboardingHandler) GetTenantByRequestID(w http.ResponseWriter, r *http.Request, id string) {
	token := r.Header.Get("Authorization")
	if token == "" {
		h.Logger.Info("Empty token received")
		return
	}
	if !validateServiceToken(w, token) {
		h.Logger.Info("Token validation failed")
		return
	}
	requests, err := h.Service.GetTenantByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	h.Logger.Info("IN handler", zap.Any("hand", requests))
	json.NewEncoder(w).Encode(requests)
}

// ApproveOnboarding approves onboarding requests
func (h *onboardingHandler) ApproveOnboarding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}
	token := r.Header.Get("Authorization")
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if !validateServiceToken(w, token) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.RequestID == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	err := h.Service.ApproveOnboarding(r.Context(), req.RequestID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Onboarding request approved, credentials sent"})
}

func validateServiceToken(w http.ResponseWriter, tokenString string) bool {
	ownerClaims, err := validateToken(tokenString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid owner token")
		return false
	}
	log.Printf("%s", ownerClaims.Role)
	if ownerClaims.Role != "superuser" {
		respondWithError(w, http.StatusForbidden, "Unauthorized: Only superusers can view/approve")
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
		return nil, errors.New("invalid token")
	}
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}
