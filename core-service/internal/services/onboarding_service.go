package services

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"github.com/mrityunjay-vashisth/core-service/internal/config"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"github.com/mrityunjay-vashisth/go-idforge/pkg/idforge"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type OnboardingServices interface {
	OnboardTenant(ctx context.Context, req models.OnboardingRequest) (string, error)
	GetTenants(ctx context.Context, status string) (interface{}, error)
	GetTenantByID(ctx context.Context, id string) (interface{}, error)
	ApproveOnboarding(ctx context.Context, requestID string) error
}

type OnboardingService struct {
	db     db.DBClientInterface
	Logger *zap.Logger
}

func NewOnboardingService(db db.DBClientInterface, logger *zap.Logger) *OnboardingService {
	return &OnboardingService{db: db, Logger: logger}
}

func (h *OnboardingService) OnboardTenant(ctx context.Context, req models.OnboardingRequest) (string, error) {
	if req.OrganizationName == "" || req.Email == "" {
		return "", errors.New("invalid request body")
	}
	requestId := idforge.GenerateWithSize(20)
	tenantId := idforge.GenerateWithSize(10)
	userId := idforge.GenerateWithSize(10)

	filter := bson.M{"email": req.Email}
	existingReq, err := h.db.Read(ctx, filter,
		db.WithDatabaseName("coredb"),
		db.WithCollectionName("onboarding_requests"))

	// First check for database errors
	if err != nil {
		h.Logger.Error("Database error", zap.Error(err))
		return "", errors.New("database error while checking for existing requests")
	}

	if existingReq != nil {
		if reqMap, ok := existingReq.(map[string]interface{}); ok && len(reqMap) > 0 {
			return "", errors.New("onboarding request already exists")
		}
	}

	existingReq, err = h.db.Read(ctx, filter,
		db.WithDatabaseName("coredb"),
		db.WithCollectionName("onboarded_tenants"))

	// First check for database errors
	if err != nil {
		h.Logger.Error("Database error", zap.Error(err))
		return "", errors.New("database error while checking for existing requests")
	}

	if existingReq != nil {
		if reqMap, ok := existingReq.(map[string]interface{}); ok && len(reqMap) > 0 {
			return "", errors.New("onboarding request already exists")
		}
	}

	dataMap := map[string]interface{}{
		"organization_name": req.OrganizationName,
		"email":             req.Email,
		"status":            "pending",
		"username":          userId,
		"tenant_id":         tenantId,
		"created_at":        time.Now().String(),
		"request_id":        requestId,
		"role":              req.Role,
		"geo_location":      "",
		"entitlements":      "",
	}
	_, err = h.db.Create(ctx, dataMap,
		db.WithDatabaseName("coredb"),
		db.WithCollectionName("onboarding_requests"))
	if err != nil {
		return "", errors.New("failed to onboard tenant")
	}

	h.Logger.Info("Onboarding request submitted", zap.String("request_id", requestId))
	return requestId, nil
}

// GetPendingRequests fetches pending onboarding requests
func (h *OnboardingService) GetTenants(ctx context.Context, status string) (interface{}, error) {
	var filter bson.M
	var dbName string
	var collectionName string
	if status == "pending" {
		filter = bson.M{"status": "pending"}
		dbName = config.Medusa_Core
		collectionName = config.Onboarding_Requests
	} else {
		filter = bson.M{"status": "active"}
		dbName = config.Medusa_Core
		collectionName = config.Onboarded_Tenants
	}
	requests, err := h.db.ReadAll(ctx, filter, db.WithDatabaseName(dbName), db.WithCollectionName(collectionName))
	if err != nil {
		return nil, errors.New("failed to fetch pending requests")
	}
	return requests, nil
}

// GetPendingRequests fetches pending onboarding requests
func (h *OnboardingService) GetTenantByID(ctx context.Context, id string) (interface{}, error) {
	filter := bson.M{"request_id": id}
	requests, err := h.db.Read(ctx, filter, db.WithDatabaseName(config.Medusa_Core), db.WithCollectionName(config.Onboarded_Tenants))
	h.Logger.Info("req", zap.Any("req", requests))
	if err != nil {
		return nil, errors.New("failed to fetch pending requests")
	}
	return requests, nil
}

func (h *OnboardingService) ApproveOnboarding(ctx context.Context, requestID string) error {
	filter := bson.M{"request_id": requestID, "status": "pending"}
	request, err := h.db.Read(ctx, filter, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))

	// Check if request is nil regardless of error
	if request == nil {
		h.Logger.Warn("Request is nil", zap.Error(err), zap.String("request_id", requestID))
		return errors.New("no pending request found with the given ID")
	}

	// Now let's explicitly handle empty map case
	requestMap, ok := request.(map[string]interface{})
	if !ok {
		h.Logger.Error("Failed to convert request to map", zap.Any("request", request))
		return errors.New("failed to process request data: type conversion failed")
	}

	// Check if the map is empty
	if len(requestMap) == 0 {
		h.Logger.Error("Request map is empty", zap.Any("request", request))
		return errors.New("failed to process request data: empty data returned")
	}

	// Create a completely new map to be safe
	newRequestMap := make(map[string]interface{})

	// Copy all existing fields
	for k, v := range requestMap {
		newRequestMap[k] = v
	}

	// Generate password
	password := generateRandomPassword()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		h.Logger.Error("Failed to hash password", zap.Error(err))
		return errors.New("failed to process credentials")
	}

	// Add new fields to the new map
	newRequestMap["password"] = string(hashedPassword)
	newRequestMap["status"] = "active"
	newRequestMap["approved_at"] = time.Now().String()

	// Insert as approved tenant
	_, err = h.db.Create(ctx, newRequestMap, db.WithDatabaseName("coredb"), db.WithCollectionName(config.Onboarding_Requests))
	if err != nil {
		h.Logger.Error("Failed to create approved tenant", zap.Error(err))
		return errors.New("failed to approve request")
	}

	// // Delete pending request
	// _, err = h.db.Delete(ctx, filter, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))
	// if err != nil {
	// 	h.Logger.Error("Failed to delete pending request", zap.Error(err))
	// 	// Continue despite error
	// }

	h.Logger.Info("Credentials prepared for",
		zap.String("email", newRequestMap["email"].(string)),
		zap.String("username", newRequestMap["username"].(string)),
		zap.String("request_id", newRequestMap["request_id"].(string)),
		zap.String("password", password))

	return nil
}

// GetPendingRequests fetches pending onboarding requests
func (h *OnboardingService) GetActiveOrg(ctx context.Context) (interface{}, error) {
	filter := bson.M{"status": "active"}
	requests, err := h.db.ReadAll(ctx, filter, db.WithDatabaseName(config.Medusa_Core), db.WithCollectionName(config.Onboarding_Requests))
	if err != nil {
		return nil, errors.New("failed to fetch pending requests")
	}
	return requests, nil
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
