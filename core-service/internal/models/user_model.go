package models

import "time"

// Base User struct (Common for all users)
type User struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Email       string    `bson:"email"`
	PhoneNumber string    `bson:"phone_number"`
	Gender      string    `bson:"gender"`
	Address     string    `bson:"address"`
	Roles       []string  `bson:"roles"` // Doctor, Nurse, Patient, Admin, etc.
	TenantID    string    `bson:"tenant_id"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}
