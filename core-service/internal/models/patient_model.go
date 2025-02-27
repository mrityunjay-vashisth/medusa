package models

import "time"

// Prescription represents detailed medication instructions
type Prescription struct {
	MedicineName string `bson:"medicine_name"`
	Dosage       string `bson:"dosage"`
	Frequency    string `bson:"frequency"`
	Duration     string `bson:"duration"`
	Notes        string `bson:"notes,omitempty"`
}

// LabReport stores uploaded test results as PDFs
type LabReport struct {
	ReportID   string    `bson:"report_id"`
	TestName   string    `bson:"test_name"`
	PDFURL     string    `bson:"pdf_url"`
	UploadedAt time.Time `bson:"uploaded_at"`
}

// Appointment represents a patient's past appointments
type Appointment struct {
	AppointmentID string    `bson:"appointment_id"`
	DoctorID      string    `bson:"doctor_id"`
	DoctorName    string    `bson:"doctor_name"`
	Date          time.Time `bson:"date"`
	Notes         string    `bson:"notes"`
}

// MedicalRecord represents a patient's medical history
type MedicalRecord struct {
	RecordID      string         `bson:"record_id"`
	Description   string         `bson:"description"`
	Diagnosis     string         `bson:"diagnosis"`
	DoctorID      string         `bson:"doctor_id"`
	DoctorName    string         `bson:"doctor_name"`
	DateRecorded  time.Time      `bson:"date_recorded"`
	Prescriptions []Prescription `bson:"prescriptions"`
	Instructions  string         `bson:"instructions"`
	LabReports    []LabReport    `bson:"lab_reports"`
	FollowUp      *time.Time     `bson:"follow_up,omitempty"`
}

// Patient represents a patient in the system
type Patient struct {
	ID             string          `bson:"_id"`
	Name           string          `bson:"name"`
	Age            int             `bson:"age"`
	Gender         string          `bson:"gender"`
	ContactNumber  string          `bson:"contact_number"`
	Email          string          `bson:"email"`
	Address        string          `bson:"address"`
	MedicalRecords []MedicalRecord `bson:"medical_records"`
	Appointments   []Appointment   `bson:"appointments"`
	TenantID       string          `bson:"tenant_id"`
}
