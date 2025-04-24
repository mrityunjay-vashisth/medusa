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
