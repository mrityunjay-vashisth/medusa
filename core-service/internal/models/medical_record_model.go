package models

import "time"

type MedicalRecord struct {
	ID           string    `bson:"_id,omitempty" json:"id,omitempty"`
	PatientID    string    `bson:"patient_id" json:"patient_id"`
	DoctorID     string    `bson:"doctor_id" json:"doctor_id"`
	Diagnosis    string    `bson:"diagnosis" json:"diagnosis"`
	Prescription string    `bson:"prescription" json:"prescription"`
	VisitDate    time.Time `bson:"visit_date" json:"visit_date"`
	Notes        string    `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
}
