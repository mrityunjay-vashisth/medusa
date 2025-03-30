package onboardingsvc

import (
	"context"
	"errors"
	"time"

	"github.com/mrityunjay-vashisth/core-service/internal/config"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"github.com/mrityunjay-vashisth/core-service/internal/services/authsvc"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

// StuckRequestRecovery periodically checks for and fixes stuck onboarding requests
type StuckRequestRecovery struct {
	db                db.DBClientInterface
	authService       authsvc.Service
	logger            *zap.Logger
	inProgressMaxAge  time.Duration // How long a request can be "in progress" before we consider it stuck
	userCreatedMaxAge time.Duration // How long a request can be in "user created" state before we consider it stuck
	ticker            *time.Ticker
	stopChan          chan struct{}
}

// NewStuckRequestRecovery creates a new recovery system
func NewStuckRequestRecovery(
	db db.DBClientInterface,
	authService authsvc.Service,
	logger *zap.Logger,
) *StuckRequestRecovery {
	return &StuckRequestRecovery{
		db:                db,
		authService:       authService,
		logger:            logger,
		inProgressMaxAge:  3 * time.Minute, // Configurable
		userCreatedMaxAge: 3 * time.Minute, // Configurable
		stopChan:          make(chan struct{}),
	}
}

// Start begins the periodic recovery process
func (r *StuckRequestRecovery) Start() {
	r.ticker = time.NewTicker(1 * time.Minute) // Run every 15 minutes

	go func() {
		for {
			select {
			case <-r.ticker.C:
				if err := r.RecoverStuckRequests(context.Background()); err != nil {
					r.logger.Error("Error recovering stuck requests", zap.Error(err))
				}
			case <-r.stopChan:
				r.ticker.Stop()
				return
			}
		}
	}()

	r.logger.Info("Stuck request recovery system started")
}

// Stop halts the recovery process
func (r *StuckRequestRecovery) Stop() {
	close(r.stopChan)
	r.logger.Info("Stuck request recovery system stopped")
}

// RecoverStuckRequests finds and fixes stuck requests
func (r *StuckRequestRecovery) RecoverStuckRequests(ctx context.Context) error {
	// Check for requests stuck in "approval_in_progress" state
	if err := r.recoverInProgressRequests(ctx); err != nil {
		return err
	}

	// Check for requests stuck in "user_created" state
	if err := r.recoverUserCreatedRequests(ctx); err != nil {
		return err
	}

	return nil
}

// recoverInProgressRequests handles requests stuck in the initial approval stage
func (r *StuckRequestRecovery) recoverInProgressRequests(ctx context.Context) error {
	cutoffTime := time.Now().Add(-r.inProgressMaxAge)

	filter := bson.M{
		"status":              models.OnboardingStatusApprovalInProgress,
		"approval_started_at": bson.M{"$lt": cutoffTime},
	}

	stuckRequests, err := r.db.ReadAll(
		ctx,
		filter,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardingRequests),
	)

	if err != nil {
		return errors.New("failed to query stuck in-progress requests: " + err.Error())
	}

	requestsArr, ok := stuckRequests.([]map[string]interface{})
	if !ok {
		return errors.New("invalid response format from database")
	}

	r.logger.Info("Found stuck in-progress requests", zap.Int("count", len(requestsArr)))

	for _, req := range requestsArr {
		requestID, _ := req["request_id"].(string)
		email, _ := req["email"].(string)
		username, _ := req["username"].(string)

		// Check if the user was actually created in auth service
		userExists, err := r.checkUserExists(ctx, email, username)
		if err != nil {
			r.logger.Error("Error checking user existence",
				zap.Error(err),
				zap.String("request_id", requestID))
			continue
		}

		if userExists {
			// User exists, move to user_created state
			now := time.Now()
			updateFilter := bson.M{"request_id": requestID}
			update := bson.M{
				"$set": bson.M{
					"status":          models.OnboardingStatusUserCreated,
					"user_created_at": now,
				},
			}

			_, err := r.db.UpdateOne(
				ctx,
				updateFilter,
				update,
				db.WithDatabaseName(config.DatabaseNames.CoreDB),
				db.WithCollectionName(config.CollectionNames.OnboardingRequests),
			)

			if err != nil {
				r.logger.Error("Failed to update stuck request to user_created state",
					zap.Error(err),
					zap.String("request_id", requestID))
				continue
			}

			r.logger.Info("Recovered stuck in-progress request - user exists",
				zap.String("request_id", requestID))
		} else {
			// User doesn't exist, revert to pending
			updateFilter := bson.M{"request_id": requestID}
			update := bson.M{
				"$set": bson.M{
					"status":        models.OnboardingStatusPending,
					"last_retry_at": time.Now(),
				},
				"$unset": bson.M{
					"approval_started_at": "",
				},
			}

			_, err := r.db.UpdateOne(
				ctx,
				updateFilter,
				update,
				db.WithDatabaseName(config.DatabaseNames.CoreDB),
				db.WithCollectionName(config.CollectionNames.OnboardingRequests),
			)

			if err != nil {
				r.logger.Error("Failed to revert stuck request to pending state",
					zap.Error(err),
					zap.String("request_id", requestID))
				continue
			}

			r.logger.Info("Recovered stuck in-progress request - reverted to pending",
				zap.String("request_id", requestID))
		}
	}

	return nil
}

// recoverUserCreatedRequests handles requests where the user was created but approval wasn't completed
func (r *StuckRequestRecovery) recoverUserCreatedRequests(ctx context.Context) error {
	cutoffTime := time.Now().Add(-r.userCreatedMaxAge)

	filter := bson.M{
		"status":          models.OnboardingStatusUserCreated,
		"user_created_at": bson.M{"$lt": cutoffTime},
	}

	stuckRequests, err := r.db.ReadAll(
		ctx,
		filter,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.OnboardingRequests),
	)

	if err != nil {
		return errors.New("failed to query stuck user-created requests: " + err.Error())
	}

	requestsArr, ok := stuckRequests.([]map[string]interface{})
	if !ok {
		return errors.New("invalid response format from database")
	}

	r.logger.Info("Found stuck user-created requests", zap.Int("count", len(requestsArr)))

	for _, req := range requestsArr {
		requestID, _ := req["request_id"].(string)

		// Complete the approval process
		updateFilter := bson.M{"request_id": requestID}

		// Create a new map for approved tenant
		approvedRequest := make(map[string]interface{})

		// Copy all existing fields except MongoDB internal _id
		for k, v := range req {
			if k != "_id" {
				approvedRequest[k] = v
			}
		}

		// Update status and timestamps
		now := time.Now()
		approvedRequest["status"] = models.OnboardingStatusActive
		approvedRequest["approved_at"] = now.String()

		// Insert into onboarded_tenants collection
		_, err = r.db.Create(
			ctx,
			approvedRequest,
			db.WithDatabaseName(config.DatabaseNames.CoreDB),
			db.WithCollectionName(config.CollectionNames.OnboardedTenants),
		)

		if err != nil {
			r.logger.Info("Failed to create approved tenant record during recovery",
				zap.Error(err),
				zap.String("request_id", requestID))
			continue
		}

		// Delete from onboarding_requests collection
		_, err = r.db.Delete(
			ctx,
			updateFilter,
			db.WithDatabaseName(config.DatabaseNames.CoreDB),
			db.WithCollectionName(config.CollectionNames.OnboardingRequests),
		)

		if err != nil {
			r.logger.Warn("Failed to delete original request after recovery approval",
				zap.Error(err),
				zap.String("request_id", requestID))
			// Continue anyway as the tenant is approved
		}

		r.logger.Info("Recovered stuck user-created request - approval completed",
			zap.String("request_id", requestID))
	}

	return nil
}

// checkUserExists verifies if a user exists in the auth service
func (r *StuckRequestRecovery) checkUserExists(ctx context.Context, email, username string) (bool, error) {
	// This would typically call your auth service's user verification endpoint
	// For now, we'll implement a simple check

	// You'd implement this based on your auth service API
	// For example:
	// resp, err := r.authService.GetClient().CheckUserExists(ctx, &authpb.CheckUserRequest{
	//     Email: email,
	//     Username: username,
	// })
	// return resp.Exists, err

	// For now, return a mock implementation
	return true, nil
}
