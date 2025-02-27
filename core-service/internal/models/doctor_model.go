package models

import "time"

// DoctorAppointment represents a doctor's appointments with patients
type DoctorAppointment struct {
	AppointmentID string    `bson:"appointment_id"`
	PatientID     string    `bson:"patient_id"`
	PatientName   string    `bson:"patient_name"`
	Date          time.Time `bson:"date"`
	Status        string    `bson:"status"` // Scheduled, Completed, Cancelled
	Notes         string    `bson:"notes,omitempty"`
}

// AvailabilitySlot represents a doctor's available slots for booking
type AvailabilitySlot struct {
	Day       string `bson:"day"`        // Monday, Tuesday, etc.
	StartTime string `bson:"start_time"` // e.g., "09:00"
	EndTime   string `bson:"end_time"`   // e.g., "13:00"
	IsBooked  bool   `bson:"is_booked"`
}

// Review represents patient feedback for a doctor
type Review struct {
	ReviewID  string    `bson:"review_id"`
	PatientID string    `bson:"patient_id"`
	Rating    float64   `bson:"rating"`
	Comment   string    `bson:"comment,omitempty"`
	Timestamp time.Time `bson:"timestamp"`
}

// Doctor represents a doctor in the system (Now properly linking other entities)
type Doctor struct {
	ID                          string              `bson:"_id"`
	Name                        string              `bson:"name"`
	Email                       string              `bson:"email"`
	PhoneNumber                 string              `bson:"phone_number"`
	Specialization              string              `bson:"specialization"`
	Qualifications              []string            `bson:"qualifications"`
	LicenseNumber               string              `bson:"license_number"`
	Certifications              []string            `bson:"certifications"`
	MedicalAssociations         []string            `bson:"medical_associations"`
	YearsOfPractice             int                 `bson:"years_of_practice"`
	LanguagesSpoken             []string            `bson:"languages_spoken"`
	ConsultationFee             float64             `bson:"consultation_fee"`
	Currency                    string              `bson:"currency"`
	ConsultationMode            []string            `bson:"consultation_mode"`
	EstimatedWaitTime           string              `bson:"estimated_wait_time"`
	PatientLimitPerDay          int                 `bson:"patient_limit_per_day"`
	HospitalAffiliations        []string            `bson:"hospital_affiliations"`
	ClinicAddress               string              `bson:"clinic_address"`
	IsFullTime                  bool                `bson:"is_full_time"`
	OnlineConsultationPlatforms []string            `bson:"online_consultation_platforms"`
	AverageRating               float64             `bson:"average_rating"`
	TotalReviews                int                 `bson:"total_reviews"`
	Reviews                     []Review            `bson:"reviews"`
	Appointments                []DoctorAppointment `bson:"appointments"`
	Availability                []AvailabilitySlot  `bson:"availability"`
	TenantID                    string              `bson:"tenant_id"`
}
