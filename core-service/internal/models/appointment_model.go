package models

import (
	"time"
)

// AppointmentStatus represents the current status of an appointment
type AppointmentStatus string

const (
	AppointmentStatusScheduled   AppointmentStatus = "scheduled"
	AppointmentStatusCompleted   AppointmentStatus = "completed"
	AppointmentStatusCancelled   AppointmentStatus = "cancelled"
	AppointmentStatusNoShow      AppointmentStatus = "no_show"
	AppointmentStatusRescheduled AppointmentStatus = "rescheduled"
)

// AppointmentType represents different types of appointments
type AppointmentType string

const (
	AppointmentTypeRoutine      AppointmentType = "routine"
	AppointmentTypeUrgent       AppointmentType = "urgent"
	AppointmentTypeFollowUp     AppointmentType = "follow_up"
	AppointmentTypeConsultation AppointmentType = "consultation"
	AppointmentTypeSpecialist   AppointmentType = "specialist"
)

// Appointment represents a scheduled appointment between a patient and a doctor
type Appointment struct {
	AppointmentID string            `json:"appointment_id" bson:"_id"`
	PatientID     string            `json:"patient_id" bson:"patient_id"`
	PatientName   string            `json:"patient_name" bson:"patient_name"`
	DoctorID      string            `json:"doctor_id" bson:"doctor_id"`
	DoctorName    string            `json:"doctor_name" bson:"doctor_name"`
	ScheduledTime time.Time         `json:"scheduled_time" bson:"scheduled_time"`
	Duration      int               `json:"duration" bson:"duration"` // Duration in minutes
	Type          AppointmentType   `json:"appointment_type" bson:"appointment_type"`
	Status        AppointmentStatus `json:"status" bson:"status"`
	Notes         string            `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedBy     string            `json:"created_by" bson:"created_by"` // User ID of the creator (receptionist)
	CreatedAt     time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" bson:"updated_at"`
	TenantID      string            `json:"tenant_id" bson:"tenant_id"`
}

// AppointmentCreateRequest is used to create a new appointment
type AppointmentCreateRequest struct {
	PatientID     string          `json:"patient_id"`
	PatientName   string          `json:"patient_name"`
	DoctorID      string          `json:"doctor_id"`
	DoctorName    string          `json:"doctor_name"`
	ScheduledTime time.Time       `json:"scheduled_time"`
	Duration      int             `json:"duration"`
	Type          AppointmentType `json:"appointment_type"`
	Notes         string          `json:"notes,omitempty"`
}

// AppointmentUpdateRequest is used to update an existing appointment
type AppointmentUpdateRequest struct {
	ScheduledTime *time.Time         `json:"scheduled_time,omitempty"`
	Duration      *int               `json:"duration,omitempty"`
	Type          *AppointmentType   `json:"appointment_type,omitempty"`
	Status        *AppointmentStatus `json:"status,omitempty"`
	Notes         *string            `json:"notes,omitempty"`
}

// AppointmentResponse is the response returned after appointment operations
type AppointmentResponse struct {
	AppointmentID string            `json:"appointment_id"`
	PatientID     string            `json:"patient_id"`
	PatientName   string            `json:"patient_name"`
	DoctorID      string            `json:"doctor_id"`
	DoctorName    string            `json:"doctor_name"`
	ScheduledTime time.Time         `json:"scheduled_time"`
	Duration      int               `json:"duration"`
	Type          AppointmentType   `json:"appointment_type"`
	Status        AppointmentStatus `json:"status"`
	Notes         string            `json:"notes,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}
