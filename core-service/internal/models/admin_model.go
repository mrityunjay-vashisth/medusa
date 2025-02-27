package models

import "time"

// AdminActivity represents actions performed by an admin
type AdminActivity struct {
	ActivityID string    `bson:"activity_id"`
	AdminID    string    `bson:"admin_id"`
	Action     string    `bson:"action"`
	Timestamp  time.Time `bson:"timestamp"`
	Details    string    `bson:"details,omitempty"`
}

// Admin represents an administrative user in the system
type Admin struct {
	ID                 string          `bson:"_id"`
	UserID             string          `bson:"user_id"` // Links to User model
	RoleLevel          string          `bson:"role_level"`
	Permissions        []string        `bson:"permissions"`
	AssignedHospitals  []string        `bson:"assigned_hospitals"`
	ManagedDepartments []string        `bson:"managed_departments"`
	Responsibilities   []string        `bson:"responsibilities"`
	ActivityLogs       []AdminActivity `bson:"activity_logs"`
	TenantID           string          `bson:"tenant_id"`
	CreatedAt          time.Time       `bson:"created_at"`
	UpdatedAt          time.Time       `bson:"updated_at"`
}
