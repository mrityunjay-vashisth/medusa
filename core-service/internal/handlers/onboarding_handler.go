package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	logger *zap.Logger
	jwtKey = []byte("my-secret-key")
)

// OnboardingHandler handles all onboarding-related requests
type OnboardingHandler struct {
	db db.DBClientInterface
}

type onboardingRequest struct {
	OrganizationName string `json:"organization_name" bson:"organization_name"`
	Email            string `json:"email" bson:"email"`
	Role             string `json:"role" bson:"role"`
}

type entityMetadata struct {
	OrganizationName string `bson:"organization_name"`
	Email            string `bson:"email"`
	Status           string `bson:"status"`
	Username         string `bson:"username"`
	TenantID         string `bson:"tenant_id"`
	CreatedAt        string `bson:"created_at"`
	RequestID        string `bson:"request_id"`
	GeoLocation      string `bson:"geo_location"`
	Entitlements     string `bson:"entitlements"`
	Role             string `bson:"role"`
}

type claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// NewOnboardingHandler initializes the onboarding subhandler
func NewOnboardingHandler(db db.DBClientInterface) *OnboardingHandler {
	return &OnboardingHandler{db: db}
}

// ServeHTTP routes requests to onboarding-specific functions
func (h *OnboardingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger = GetLoggerFromContext(r)
	vars := mux.Vars(r)
	logger.Info("Onboarding API", zap.String("subpath", vars["subpath"]))
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
	var req onboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.OrganizationName == "" || req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	requestId := uuid.New().String()
	tenantId := uuid.New().String()
	userId := uuid.New().String()

	filter := bson.M{"email": req.Email}
	existingReq, err := h.db.Read(r.Context(), filter, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))
	if err == nil && existingReq != nil {
		respondWithError(w, http.StatusConflict, "Onboarding request already exists")
		return
	}

	reqData := &entityMetadata{
		OrganizationName: req.OrganizationName,
		Email:            req.Email,
		Status:           "pending",
		Username:         userId,
		TenantID:         tenantId,
		CreatedAt:        time.Now().String(),
		RequestID:        requestId,
		Role:             req.Role,
		GeoLocation:      "",
		Entitlements:     "",
	}

	dataMap := map[string]interface{}{
		"organization_name": reqData.OrganizationName,
		"email":             reqData.Email,
		"status":            reqData.Status,
		"username":          reqData.Username,
		"tenant_id":         reqData.TenantID,
		"created_at":        reqData.CreatedAt,
		"request_id":        reqData.RequestID,
		"geo_location":      reqData.GeoLocation,
		"entitlements":      reqData.Entitlements,
		"role":              reqData.Role,
	}
	_, err = h.db.Create(r.Context(), dataMap, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to onboard tenant")
		return
	}

	logger.Info("Onboarding request submitted", zap.String("request_id", requestId))
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

	filter := bson.M{"status": "pending"}
	requests, err := h.db.ReadAll(r.Context(), filter, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch pending requests")
		return
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

	filter := bson.M{"request_id": req.RequestID, "status": "pending"}
	request, err := h.db.Read(r.Context(), filter, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))
	if err != nil || request == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	requestMap, ok := request.(map[string]interface{})
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Failed to process request data")
		return
	}

	password := generateRandomPassword()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	requestMap["password"] = string(hashedPassword)
	requestMap["status"] = "active"

	// Insert approved request
	_, err = h.db.Create(context.Background(), requestMap, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarded_tenants"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to approve request")
		return
	}

	// Delete the old pending request
	_, err = h.db.Delete(context.Background(), filter, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to approve request")
		return
	}

	log.Printf("Sending admin credentials: Email: %s, User ID: %s, Password: %s", requestMap["email"], requestMap["userid"], password)
	json.NewEncoder(w).Encode(map[string]string{"message": "Onboarding request approved, credentials sent"})
}

func GetLoggerFromContext(r *http.Request) *zap.Logger {
	logger, ok := r.Context().Value(middleware.LoggerKey).(*zap.Logger)
	if !ok {
		return zap.NewNop() // Return a no-op logger if none found
	}
	return logger
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

func validateToken(tokenString string) (*claims, error) {
	claims := &claims{}
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

func generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
