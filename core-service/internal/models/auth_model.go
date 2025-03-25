package models

type LoginRequest struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"username"`
}

// RegisterRequest represents user registration data
type AuthRegisterRequest struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Email    string `json:"email" bson:"email"`
	Name     string `json:"name" bson:"name"`
	Role     string `json:"role" bson:"role"`
	TenantId string `json:"tenantid" bson:"tenant_id"`
}
