package services

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type OnboardingServices interface {
	OnboardTenant(ctx context.Context, req models.OnboardingRequest) (string, error)
	GetPendingRequests(ctx context.Context) (interface{}, error)
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
	requestId := uuid.New().String()
	tenantId := uuid.New().String()
	userId := uuid.New().String()

	filter := bson.M{"email": req.Email}
	existingReq, err := h.db.Read(ctx, filter, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))
	h.Logger.Info("response", zap.Any("ex", existingReq))
	// First check for database errors
	if err != nil {
		h.Logger.Error("Database error", zap.Error(err))
		return "", errors.New("database error while checking for existing requests")
	}

	// Then check if a request exists
	if existingReq != nil {
		// Need to determine how your DB client represents "no results"
		// For MongoDB, it might return a non-nil map with no entries
		if reqMap, ok := existingReq.(map[string]interface{}); ok && len(reqMap) > 0 {
			return "", errors.New("onboarding request already exists")
		}
	}

	reqData := &models.EntityMetadata{
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
	_, err = h.db.Create(ctx, dataMap, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))
	if err != nil {
		return "", errors.New("failed to onboard tenant")
	}

	h.Logger.Info("Onboarding request submitted", zap.String("request_id", requestId))
	return requestId, nil
}

// GetPendingRequests fetches pending onboarding requests
func (h *OnboardingService) GetPendingRequests(ctx context.Context) (interface{}, error) {
	filter := bson.M{"status": "pending"}
	requests, err := h.db.ReadAll(ctx, filter, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))
	if err != nil {
		return nil, errors.New("failed to fetch pending requests")
	}
	return requests, nil
}

// ApproveOnboarding approves onboarding requests
func (h *OnboardingService) ApproveOnboarding(ctx context.Context, requestID string) error {
	filter := bson.M{"request_id": requestID, "status": "pending"}
	request, err := h.db.Read(ctx, filter, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))
	if err != nil || request == nil {
		return errors.New("invalid request ID")
	}
	requestMap, ok := request.(map[string]interface{})
	if !ok {
		return errors.New("failed to process request data")
	}

	password := generateRandomPassword()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	requestMap["password"] = string(hashedPassword)
	requestMap["status"] = "active"

	// Insert approved request
	_, err = h.db.Create(context.Background(), requestMap, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarded_tenants"))
	if err != nil {
		return errors.New("failed to approve request")
	}

	// Delete the old pending request
	_, err = h.db.Delete(context.Background(), filter, db.WithDatabaseName("coredb"), db.WithCollectionName("onboarding_requests"))
	if err != nil {
		return errors.New("failed to approve request")
	}

	h.Logger.Info("Sending admin credentials:", zap.String("Email", requestMap["email"].(string)), zap.String("userid", requestMap["userid"].(string)))
	return nil
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
