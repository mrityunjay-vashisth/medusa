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

// OnboardingHandler handles all onboarding-related requests
type OnboardingHandler struct {
	Service services.OnboardingServices
	Logger  *zap.Logger
}

func NewOnboardingHandler(service services.OnboardingServices, logger *zap.Logger) *OnboardingHandler {
	return &OnboardingHandler{Service: service, Logger: logger}
}

// ServeHTTP routes requests to onboarding-specific functions
func (h *OnboardingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.Logger.Info("Onboarding API", zap.String("subpath", vars["subpath"]))
	subPath := vars["subpath"]

	switch subPath {
	case "pending":
		h.GetPendingRequests(w, r)
	case "approve":
		h.ApproveOnboarding(w, r)
	case "onboard":
		h.OnboardTenant(w, r)
	default:
		respondWithError(w, http.StatusNotFound, "Invalid Onboarding API")
	}
}

// OnboardTenant handles onboarding requests
func (h *OnboardingHandler) OnboardTenant(w http.ResponseWriter, r *http.Request) {
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
func (h *OnboardingHandler) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
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
	requests, err := h.Service.GetPendingRequests(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	json.NewEncoder(w).Encode(requests)
}

// ApproveOnboarding approves onboarding requests
func (h *OnboardingHandler) ApproveOnboarding(w http.ResponseWriter, r *http.Request) {
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

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
