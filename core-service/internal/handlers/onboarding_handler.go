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
	"github.com/mrityunjay-vashisth/core-service/internal/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	logger *zap.Logger
	jwtKey = []byte("my-secret-key")
)

// OnboardingHandler handles all onboarding-related requests
type OnboardingHandler struct {
	db *mongo.Client
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
func NewOnboardingHandler(db *mongo.Client) *OnboardingHandler {
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
		http.Error(w, "Invalid Onboarding API", http.StatusNotFound)
	}
}

// OnboardTenant handles onboarding requests
func (h *OnboardingHandler) OnboardTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var req onboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.OrganizationName == "" || req.Email == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	requestId := uuid.New().String()
	tenantId := uuid.New().String()
	userId := uuid.New().String()

	collection := h.db.Database("coredb").Collection("onboarding_requests")
	existingReq := collection.FindOne(r.Context(), map[string]string{"email": req.Email})
	if existingReq.Err() == nil {
		http.Error(w, "Onboarding request already exists", http.StatusConflict)
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
	_, err := collection.InsertOne(r.Context(), reqData)
	if err != nil {
		http.Error(w, "Failed to onboard tenant", http.StatusInternalServerError)
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
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !validateServiceToken(w, token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	collection := h.db.Database("coredb").Collection("onboarding_requests")
	cursor, err := collection.Find(r.Context(), bson.M{"status": "pending"})
	if err != nil {
		http.Error(w, "Failed to fetch pending requests", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(r.Context())
	var requests []bson.M
	if err := cursor.All(r.Context(), &requests); err != nil {
		http.Error(w, "Failed to fetch pending requests", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(requests)
}

// ApproveOnboarding approves onboarding requests
func (h *OnboardingHandler) ApproveOnboarding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !validateServiceToken(w, token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.RequestID == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	collection := h.db.Database("coredb").Collection("onboarding_requests")
	entitlementsCollection := h.db.Database("coredb").Collection("onboarded_tenants")
	var request bson.M
	err := collection.FindOne(r.Context(), bson.M{"request_id": req.RequestID, "status": "pending"}).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}
	password := generateRandomPassword()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	request["password"] = string(hashedPassword)
	request["status"] = "active"
	_, err = entitlementsCollection.InsertOne(context.Background(), request)
	if err != nil {
		http.Error(w, "Failed to approve request", http.StatusInternalServerError)
		return
	}
	_, _ = collection.DeleteOne(context.Background(), bson.M{"request_id": req.RequestID})
	log.Printf("Sending admin credentials: Email: %s, User ID: %s, Password: %s", request["email"], request["userid"], password)
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
		http.Error(w, "Invalid owner token", http.StatusUnauthorized)
		return false
	}
	log.Printf("%s", ownerClaims.Role)
	if ownerClaims.Role != "superuser" {
		http.Error(w, "Unauthorized: Only service owners can view/approve", http.StatusForbidden)
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
