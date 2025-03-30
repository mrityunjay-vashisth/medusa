package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type OnboardingStatus string

const (
	OnboardingStatusPending            OnboardingStatus = "pending"
	OnboardingStatusApprovalInProgress OnboardingStatus = "approval_in_progress"
	OnboardingStatusUserCreated        OnboardingStatus = "user_created"
	OnboardingStatusActive             OnboardingStatus = "active"
	OnboardingStatusFailed             OnboardingStatus = "failed"
)

// OnboardingRequest represents a new onboarding request
type OnboardingRequest struct {
	OrganizationName   string           `json:"organization_name" bson:"organization_name"`
	Email              string           `json:"email" bson:"email"`
	Role               string           `json:"role" bson:"role"`
	Address            string           `json:"address" bson:"address"`                         // Add this
	PhoneNumber        string           `json:"phone_number" bson:"phone_number"`               // Add this
	BusinessIdentifier string           `json:"business_identifier" bson:"business_identifier"` // Add this (tax ID, registration number, etc.)
	Status             OnboardingStatus `json:"status" bson:"status"`
	ApprovalStartedAt  *time.Time       `json:"approval_started_at,omitempty" bson:"approval_started_at,omitempty"`
	UserCreatedAt      *time.Time       `json:"user_created_at,omitempty" bson:"user_created_at,omitempty"`
	ApprovedAt         *time.Time       `json:"approved_at,omitempty" bson:"approved_at,omitempty"`
	FailureReason      string           `json:"failure_reason,omitempty" bson:"failure_reason,omitempty"`
	RetryCount         int              `json:"retry_count,omitempty" bson:"retry_count,omitempty"`
	LastRetryAt        *time.Time       `json:"last_retry_at,omitempty" bson:"last_retry_at,omitempty"`
}

// EntityMetadata represents the structure of an onboarding request stored in MongoDB
type EntityMetadata struct {
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

// UserClaims represents JWT claims extracted from a token
type UserClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}
