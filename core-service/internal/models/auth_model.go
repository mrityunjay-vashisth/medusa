package models

type LoginRequest struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"username"`
}

type RegisterRequest struct {
	Name     string
	Email    string
	Username string
	Password string
	Gender   string
	Role     string
	TenantID string
}
