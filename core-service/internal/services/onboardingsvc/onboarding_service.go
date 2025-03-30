package onboardingsvc

import (
	"context"
	"errors"
	"time"

	"github.com/mrityunjay-vashisth/core-service/internal/config"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"github.com/mrityunjay-vashisth/go-idforge/pkg/idforge"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type Service interface {
	OnboardTenant(ctx context.Context, req models.OnboardingRequest) (string, error)
	GetTenants(ctx context.Context, status string) (interface{}, error)
	GetTenantByID(ctx context.Context, id string) (interface{}, error)
	BeginApproval(ctx context.Context, requestID string) (map[string]interface{}, error)
	MarkUserCreated(ctx context.Context, requestID string) error
	CompleteApproval(ctx context.Context, requestID string) error
	MarkApprovalFailed(ctx context.Context, requestID string, reason string) error
	RevertToRetriable(ctx context.Context, requestID string) error
}

type onboardingService struct {
	db          db.DBClientInterface
	Logger      *zap.Logger
	svcRegistry registry.ServiceRegistry
}

func NewService(db db.DBClientInterface, registry registry.ServiceRegistry, logger *zap.Logger) Service {
	return &onboardingService{
		db:          db,
		svcRegistry: registry,
		Logger:      logger,
	}
}

func (h *onboardingService) OnboardTenant(ctx context.Context, req models.OnboardingRequest) (string, error) {
	requestId := idforge.GenerateWithSize(20)
	tenantId := idforge.GenerateWithSize(10)
	username := idforge.GenerateWithSize(10)

	filter := bson.M{"email": req.Email}
	existingReq, err := h.db.Read(ctx, filter,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardingRequests))

	// First check for database errors
	if err != nil {
		return "", errors.New("database error while checking for existing requests")
	}

	if existingReq != nil {
		if reqMap, ok := existingReq.(map[string]interface{}); ok && len(reqMap) > 0 {
			return "", errors.New("onboarding request already exists")
		}
	}

	existingReq, err = h.db.Read(ctx, filter,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardedTenants))

	// First check for database errors
	if err != nil {
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
		"tenant_id":         tenantId,
		"created_at":        time.Now().String(),
		"request_id":        requestId,
		"role":              req.Role,
		"username":          username,
		"geo_location":      "",
		"entitlements":      "",
	}
	_, err = h.db.Create(ctx, dataMap,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardingRequests))
	if err != nil {
		return "", errors.New("failed to onboard tenant")
	}
	return requestId, nil
}

// GetPendingRequests fetches pending onboarding requests
func (h *onboardingService) GetTenants(ctx context.Context, status string) (interface{}, error) {
	var filter bson.M
	var dbName string
	var collectionName string
	if status == "pending" {
		filter = bson.M{"status": "pending"}
		dbName = config.DatabaseNames.CoreDB
		collectionName = config.CollectionNames.OnboardingRequests
	} else {
		filter = bson.M{"status": "active"}
		dbName = config.DatabaseNames.CoreDB
		collectionName = config.CollectionNames.OnboardedTenants
	}
	requests, err := h.db.ReadAll(ctx, filter,
		db.WithDatabaseName(dbName),
		db.WithCollectionName(collectionName))
	if err != nil {
		h.Logger.Info("Error reading pending", zap.String("err", err.Error()))
		return nil, errors.New("failed to fetch pending requests")
	}
	return requests, nil
}

// GetPendingRequests fetches pending onboarding requests
func (h *onboardingService) GetTenantByID(ctx context.Context, id string) (interface{}, error) {
	filter := bson.M{"request_id": id}
	requests, err := h.db.Read(ctx, filter,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardingRequests))

	if err != nil {
		return nil, errors.New("failed to fetch pending requests")
	}

	// Now let's explicitly handle empty map case
	requestMap, ok := requests.(map[string]interface{})
	if !ok {
		return nil, errors.New("failed to process request data: type conversion failed")
	}
	// Check if the map is empty
	if len(requestMap) == 0 {
		return nil, errors.New("request id not approved, please contact support")
	}
	return requests, nil
}

// BeginApproval marks an onboarding request as "in progress"
func (h *onboardingService) BeginApproval(ctx context.Context, requestID string) (map[string]interface{}, error) {
	now := time.Now()
	filter := bson.M{"request_id": requestID, "status": models.OnboardingStatusPending}
	update := bson.M{
		"$set": bson.M{
			"status":              models.OnboardingStatusApprovalInProgress,
			"approval_started_at": now,
		},
	}

	// Update the document
	result, err := h.db.UpdateOne(
		ctx,
		filter,
		update,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardingRequests),
	)

	if err != nil {
		h.Logger.Error("Failed to begin approval process",
			zap.Error(err),
			zap.String("request_id", requestID))
		return nil, errors.New("database error while beginning approval")
	}

	if result == 0 {
		// Document wasn't updated - might not exist or not be in pending state
		h.Logger.Warn("No pending request found for approval",
			zap.String("request_id", requestID))
		return nil, errors.New("no pending request found with the given ID")
	}

	// Get the updated request data
	request, err := h.GetTenantByID(ctx, requestID)
	if err != nil {
		h.Logger.Error("Failed to retrieve request after beginning approval",
			zap.Error(err),
			zap.String("request_id", requestID))
		return nil, errors.New("failed to retrieve request data")
	}

	return request.(map[string]interface{}), nil
}

// MarkUserCreated updates the request to indicate the user was created
func (h *onboardingService) MarkUserCreated(ctx context.Context, requestID string) error {
	now := time.Now()
	filter := bson.M{"request_id": requestID, "status": models.OnboardingStatusApprovalInProgress}
	update := bson.M{
		"$set": bson.M{
			"status":          models.OnboardingStatusUserCreated,
			"user_created_at": now,
		},
	}

	// Update the document
	result, err := h.db.UpdateOne(
		ctx,
		filter,
		update,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardingRequests),
	)

	if err != nil {
		h.Logger.Error("Failed to mark user as created",
			zap.Error(err),
			zap.String("request_id", requestID))
		return errors.New("database error while updating user creation status")
	}

	if result == 0 {
		// Document wasn't updated
		h.Logger.Warn("No in-progress request found for marking user created",
			zap.String("request_id", requestID))
		return errors.New("no in-progress request found with the given ID")
	}

	return nil
}

func (h *onboardingService) CompleteApproval(ctx context.Context, requestID string) error {
	// Get the request data with user_created status
	filter := bson.M{"request_id": requestID, "status": models.OnboardingStatusUserCreated}
	request, err := h.db.Read(
		ctx,
		filter,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardingRequests),
	)

	if err != nil {
		h.Logger.Error("Failed to retrieve request for completion",
			zap.Error(err),
			zap.String("request_id", requestID))
		return errors.New("database error while retrieving request")
	}

	if request == nil {
		h.Logger.Warn("No user-created request found for completion",
			zap.String("request_id", requestID))
		return errors.New("no user-created request found with the given ID")
	}

	// Convert to map if not already
	requestMap, ok := request.(map[string]interface{})
	if !ok {
		h.Logger.Error("Invalid request data format",
			zap.Any("request", request))
		return errors.New("invalid request data format")
	}

	// Create a new map for approved tenant
	now := time.Now()
	approvedRequest := make(map[string]interface{})

	// Copy all existing fields except MongoDB internal _id
	for k, v := range requestMap {
		if k != "_id" {
			approvedRequest[k] = v
		}
	}

	// Update status and timestamps
	approvedRequest["status"] = models.OnboardingStatusActive
	approvedRequest["approved_at"] = now.String()

	// Insert into onboarded_tenants collection
	_, err = h.db.Create(
		ctx,
		approvedRequest,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardedTenants),
	)

	if err != nil {
		h.Logger.Error("Failed to create approved tenant record",
			zap.Error(err),
			zap.String("request_id", requestID))
		return errors.New("failed to complete approval process")
	}

	// Delete from onboarding_requests collection
	_, err = h.db.Delete(
		ctx,
		filter,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardingRequests),
	)

	if err != nil {
		h.Logger.Warn("Failed to delete original request after approval",
			zap.Error(err),
			zap.String("request_id", requestID))
		// We won't return an error here as the tenant is already approved
		// This will be handled by cleanup processes
	}

	h.Logger.Info("Onboarding approval completed successfully",
		zap.String("request_id", requestID),
		zap.String("tenant_id", approvedRequest["tenant_id"].(string)),
		zap.String("email", approvedRequest["email"].(string)),
	)

	return nil
}

// MarkApprovalFailed records a failure during the approval process
func (h *onboardingService) MarkApprovalFailed(ctx context.Context, requestID string, reason string) error {
	now := time.Now()
	filter := bson.M{"request_id": requestID}
	update := bson.M{
		"$set": bson.M{
			"status":         models.OnboardingStatusFailed,
			"failure_reason": reason,
			"failed_at":      now,
		},
		"$inc": bson.M{
			"retry_count": 1,
		},
	}

	// Update the document
	_, err := h.db.UpdateOne(
		ctx,
		filter,
		update,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardingRequests),
	)

	if err != nil {
		h.Logger.Error("Failed to mark approval as failed",
			zap.Error(err),
			zap.String("request_id", requestID))
		return errors.New("database error while recording failure")
	}

	return nil
}

// RevertToRetriable reverts a failed request back to pending status for retry
func (h *onboardingService) RevertToRetriable(ctx context.Context, requestID string) error {
	now := time.Now()
	filter := bson.M{
		"request_id": requestID,
		"status": bson.M{"$in": []string{
			string(models.OnboardingStatusApprovalInProgress),
			string(models.OnboardingStatusUserCreated),
			string(models.OnboardingStatusFailed),
		}},
	}

	update := bson.M{
		"$set": bson.M{
			"status":        models.OnboardingStatusPending,
			"last_retry_at": now,
		},
	}

	// Update the document
	result, err := h.db.UpdateOne(
		ctx,
		filter,
		update,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardingRequests),
	)

	if err != nil {
		h.Logger.Error("Failed to revert request to pending status",
			zap.Error(err),
			zap.String("request_id", requestID))
		return errors.New("database error while reverting status")
	}

	if result == 0 {
		h.Logger.Warn("No request found for reversion",
			zap.String("request_id", requestID))
		return errors.New("no eligible request found with the given ID")
	}

	return nil
}

// func (h *onboardingService) ApproveOnboarding(ctx context.Context, requestID string) error {
// 	filter := bson.M{"request_id": requestID, "status": "pending"}
// 	request, err := h.db.Read(ctx, filter,
// 		db.WithDatabaseName(config.DatabaseNames.CoreDB),
// 		db.WithCollectionName(config.CollectionNames.OnboardingRequests))

// 	// Check if request is nil regardless of error
// 	if request == nil {
// 		h.Logger.Warn("Request is nil", zap.Error(err), zap.String("request_id", requestID))
// 		return errors.New("no pending request found with the given ID")
// 	}

// 	// Now let's explicitly handle empty map case
// 	requestMap, ok := request.(map[string]interface{})
// 	if !ok {
// 		h.Logger.Info("Failed to convert request to map", zap.Any("request", request))
// 		return errors.New("failed to process request data: type conversion failed")
// 	}

// 	// Check if the map is empty
// 	if len(requestMap) == 0 {
// 		h.Logger.Info("Request map is empty", zap.Any("request", request))
// 		return errors.New("failed to process request data: empty data returned")
// 	}

// 	// Create a completely new map to be safe
// 	newRequestMap := make(map[string]interface{})

// 	// Copy all existing fields
// 	for k, v := range requestMap {
// 		if k != "_id" {
// 			newRequestMap[k] = v
// 		}
// 	}

// 	newRequestMap["status"] = "active"
// 	newRequestMap["approved_at"] = time.Now().String()

// 	// Insert as approved tenant
// 	_, err = h.db.Create(ctx, newRequestMap,
// 		db.WithDatabaseName(config.DatabaseNames.CoreDB),
// 		db.WithCollectionName(config.CollectionNames.OnboardedTenants))
// 	if err != nil {
// 		h.Logger.Info("Failed to create approved tenant", zap.Error(err))
// 		return err
// 	}

// 	// Delete pending request
// 	_, err = h.db.Delete(ctx, filter,
// 		db.WithDatabaseName(config.DatabaseNames.CoreDB),
// 		db.WithCollectionName(config.CollectionNames.OnboardingRequests))
// 	if err != nil {
// 		h.Logger.Info("Failed to delete pending request", zap.Error(err))
// 		// Continue despite error
// 	}

// 	h.Logger.Info("Approved for",
// 		zap.String("email", newRequestMap["email"].(string)),
// 		zap.String("request_id", newRequestMap["request_id"].(string)),
// 	)

// 	return nil
// }

// GetPendingRequests fetches pending onboarding requests
func (h *onboardingService) GetActiveOrg(ctx context.Context) (interface{}, error) {
	filter := bson.M{"status": "active"}
	requests, err := h.db.ReadAll(ctx, filter,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardedTenants))
	if err != nil {
		return nil, errors.New("failed to fetch pending requests")
	}
	return requests, nil
}
