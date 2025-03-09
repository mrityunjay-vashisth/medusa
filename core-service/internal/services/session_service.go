package services

import (
	"context"
	"errors"
	"time"

	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type SessionServicesInterface interface {
	CreateSession(claims *models.UserClaims) (string, error)
	// GetSession(token string) (*models.Session, error)
	// RefreshSession()
	// DeleteSession()
}

type SessionService struct {
	db     db.DBClientInterface
	Logger *zap.Logger
}

func NewSessionService(db db.DBClientInterface, logger *zap.Logger) SessionServicesInterface {
	return &SessionService{db: db, Logger: logger}
}

func (s *SessionService) CreateSession(claims *models.UserClaims) (string, error) {
	session := &models.Session{
		SessionID: "1",
		UserID:    claims.Username,
		Role:      claims.Role,
		TenantID:  claims.TenantID,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	sessionMap := map[string]interface{}{
		"session_id": session.SessionID,
		"username":   session.UserID,
		"tenant_id":  session.TenantID,
		"expires_at": session.ExpiresAt,
		"role":       session.Role,
	}

	_, err := s.db.Create(context.TODO(), sessionMap, db.WithDatabaseName("coredb"), db.WithCollectionName("session_store"))
	if err != nil {
		return "", errors.New("failed to onboard tenant")
	}

	return session.SessionID, nil
}

func (s *SessionService) GetSession(token string) (*models.Session, error) {
	var session models.Session

	filter := bson.M{"_id": token}
	resp, err := s.db.Read(context.TODO(), filter, db.WithDatabaseName("coredb"), db.WithCollectionName("session_store"))
	if err == nil && resp != nil {
		return nil, errors.New("onboarding request already exists")
	}

	respData, err := bson.Marshal(resp)
	if err != nil {
		return nil, err
	}
	err = bson.Unmarshal(respData, &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}
