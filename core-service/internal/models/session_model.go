package models

import "time"

type Session struct {
	SessionID string    `bson:"_id"`
	UserID    string    `bson:"user_id"`
	Role      string    `bson:"role"`
	TenantID  string    `bson:"tenant_id"`
	ExpiresAt time.Time `bson:"expires_at"`
}
