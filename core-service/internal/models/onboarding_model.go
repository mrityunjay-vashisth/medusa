package models

import "github.com/golang-jwt/jwt/v5"

// OnboardingRequest represents a new onboarding request
type OnboardingRequest struct {
	OrganizationName string `json:"organization_name" bson:"organization_name"`
	Email            string `json:"email" bson:"email"`
	Role             string `json:"role" bson:"role"`
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
