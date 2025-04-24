package receptionsvc

import (
	"context"
	"errors"
	"time"

	"github.com/mrityunjay-vashisth/core-service/internal/config"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"github.com/mrityunjay-vashisth/go-idforge/pkg/idforge"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type Service interface {
	// Appointment management
	CreateAppointment(ctx context.Context, req models.AppointmentCreateRequest, tenantID string, createdBy string) (*models.AppointmentResponse, error)
	GetAppointmentByID(ctx context.Context, appointmentID string, tenantID string) (*models.AppointmentResponse, error)
	UpdateAppointment(ctx context.Context, appointmentID string, req models.AppointmentUpdateRequest, tenantID string) (*models.AppointmentResponse, error)
	CancelAppointment(ctx context.Context, appointmentID string, reason string, tenantID string) (*models.AppointmentResponse, error)
	ListAppointments(ctx context.Context, filters map[string]interface{}, tenantID string) ([]models.AppointmentResponse, error)

	// Availability checking
	GetDoctorAvailability(ctx context.Context, doctorID string, date time.Time, tenantID string) ([]map[string]interface{}, error)
}

type receptionService struct {
	db          db.DBClientInterface
	logger      *zap.Logger
	svcRegistry registry.ServiceRegistry
}

func NewService(db db.DBClientInterface, registry registry.ServiceRegistry, logger *zap.Logger) Service {
	return &receptionService{
		db:          db,
		svcRegistry: registry,
		logger:      logger,
	}
}

// CreateAppointment books a new appointment for a patient
func (s *receptionService) CreateAppointment(ctx context.Context, req models.AppointmentCreateRequest, tenantID string, createdBy string) (*models.AppointmentResponse, error) {
	// First, check if the doctor is available at the requested time
	isAvailable, err := s.checkDoctorAvailability(ctx, req.DoctorID, req.ScheduledTime, req.Duration, tenantID)
	if err != nil {
		return nil, err
	}

	if !isAvailable {
		return nil, errors.New("doctor is not available at the requested time")
	}

	// Generate a unique ID for the appointment
	appointmentID := idforge.GenerateWithSize(16)
	now := time.Now()

	// Create appointment object
	appointment := models.Appointment{
		AppointmentID: appointmentID,
		PatientID:     req.PatientID,
		PatientName:   req.PatientName,
		DoctorID:      req.DoctorID,
		DoctorName:    req.DoctorName,
		ScheduledTime: req.ScheduledTime,
		Duration:      req.Duration,
		Type:          req.Type,
		Status:        models.AppointmentStatusScheduled,
		Notes:         req.Notes,
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
		TenantID:      tenantID,
	}

	// Convert to map for database insertion
	appointmentMap := map[string]interface{}{
		"_id":              appointment.AppointmentID,
		"patient_id":       appointment.PatientID,
		"patient_name":     appointment.PatientName,
		"doctor_id":        appointment.DoctorID,
		"doctor_name":      appointment.DoctorName,
		"scheduled_time":   appointment.ScheduledTime,
		"duration":         appointment.Duration,
		"appointment_type": appointment.Type,
		"status":           appointment.Status,
		"notes":            appointment.Notes,
		"created_by":       appointment.CreatedBy,
		"created_at":       appointment.CreatedAt,
		"updated_at":       appointment.UpdatedAt,
		"tenant_id":        appointment.TenantID,
	}

	// Insert into database
	_, err = s.db.Create(
		ctx,
		appointmentMap,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.Appointments),
	)
	if err != nil {
		s.logger.Error("Failed to create appointment", zap.Error(err))
		return nil, errors.New("failed to create appointment in database")
	}

	// Create response
	response := &models.AppointmentResponse{
		AppointmentID: appointment.AppointmentID,
		PatientID:     appointment.PatientID,
		PatientName:   appointment.PatientName,
		DoctorID:      appointment.DoctorID,
		DoctorName:    appointment.DoctorName,
		ScheduledTime: appointment.ScheduledTime,
		Duration:      appointment.Duration,
		Type:          appointment.Type,
		Status:        appointment.Status,
		Notes:         appointment.Notes,
		CreatedAt:     appointment.CreatedAt,
	}

	return response, nil
}

// GetAppointmentByID retrieves an appointment by its ID
func (s *receptionService) GetAppointmentByID(ctx context.Context, appointmentID string, tenantID string) (*models.AppointmentResponse, error) {
	// Prepare filter
	filter := bson.M{
		"_id":       appointmentID,
		"tenant_id": tenantID,
	}

	// Retrieve from database
	result, err := s.db.Read(
		ctx,
		filter,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.Appointments),
	)
	if err != nil {
		s.logger.Error("Failed to retrieve appointment", zap.Error(err))
		return nil, errors.New("failed to retrieve appointment from database")
	}

	// Check if appointment exists
	if result == nil {
		return nil, errors.New("appointment not found")
	}

	// Convert result to map
	appointmentMap, ok := result.(map[string]interface{})
	if !ok {
		s.logger.Error("Failed to convert appointment to map", zap.Any("result", result))
		return nil, errors.New("invalid appointment data format")
	}

	// Create response
	response, err := s.mapToAppointmentResponse(appointmentMap)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// UpdateAppointment updates an existing appointment
func (s *receptionService) UpdateAppointment(ctx context.Context, appointmentID string, req models.AppointmentUpdateRequest, tenantID string) (*models.AppointmentResponse, error) {
	// First, retrieve the existing appointment
	existingAppointment, err := s.GetAppointmentByID(ctx, appointmentID, tenantID)
	if err != nil {
		return nil, err
	}

	// Prepare update object
	updateFields := bson.M{
		"updated_at": time.Now(),
	}

	// Only add fields that are provided in the request
	if req.ScheduledTime != nil {
		// Check if the new time is available
		isAvailable, err := s.checkDoctorAvailability(ctx, existingAppointment.DoctorID, *req.ScheduledTime, existingAppointment.Duration, tenantID)
		if err != nil {
			return nil, err
		}

		if !isAvailable {
			return nil, errors.New("doctor is not available at the requested time")
		}
		updateFields["scheduled_time"] = *req.ScheduledTime
	}

	if req.Duration != nil {
		updateFields["duration"] = *req.Duration
	}

	if req.Type != nil {
		updateFields["appointment_type"] = *req.Type
	}

	if req.Status != nil {
		updateFields["status"] = *req.Status
	}

	if req.Notes != nil {
		updateFields["notes"] = *req.Notes
	}

	// Prepare filter
	filter := bson.M{
		"_id":       appointmentID,
		"tenant_id": tenantID,
	}

	// Update in database
	_, err = s.db.UpdateOne(
		ctx,
		filter,
		bson.M{"$set": updateFields},
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.Appointments),
	)
	if err != nil {
		s.logger.Error("Failed to update appointment", zap.Error(err))
		return nil, errors.New("failed to update appointment in database")
	}

	// Retrieve updated appointment
	return s.GetAppointmentByID(ctx, appointmentID, tenantID)
}

// CancelAppointment cancels an existing appointment
func (s *receptionService) CancelAppointment(ctx context.Context, appointmentID string, reason string, tenantID string) (*models.AppointmentResponse, error) {
	// Prepare filter
	filter := bson.M{
		"_id":       appointmentID,
		"tenant_id": tenantID,
	}

	// Prepare update
	update := bson.M{
		"$set": bson.M{
			"status":     models.AppointmentStatusCancelled,
			"notes":      reason,
			"updated_at": time.Now(),
		},
	}

	// Update in database
	_, err := s.db.UpdateOne(
		ctx,
		filter,
		update,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.Appointments),
	)
	if err != nil {
		s.logger.Error("Failed to cancel appointment", zap.Error(err))
		return nil, errors.New("failed to cancel appointment in database")
	}

	// Retrieve updated appointment
	return s.GetAppointmentByID(ctx, appointmentID, tenantID)
}

// ListAppointments returns appointments based on provided filters
func (s *receptionService) ListAppointments(ctx context.Context, filters map[string]interface{}, tenantID string) ([]models.AppointmentResponse, error) {
	// Add tenant ID to filters
	filters["tenant_id"] = tenantID

	// Convert string dates to time.Time if present
	if dateFromStr, ok := filters["date_from"].(string); ok {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err == nil {
			delete(filters, "date_from")
			filters["scheduled_time"] = bson.M{"$gte": dateFrom}
		}
	}

	if dateToStr, ok := filters["date_to"].(string); ok {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err == nil {
			delete(filters, "date_to")
			// If we already have a scheduled_time filter, append to it
			if scheduleFilter, exists := filters["scheduled_time"].(bson.M); exists {
				scheduleFilter["$lte"] = dateTo.AddDate(0, 0, 1) // Include the whole day
			} else {
				filters["scheduled_time"] = bson.M{"$lte": dateTo.AddDate(0, 0, 1)}
			}
		}
	}

	// Retrieve from database
	results, err := s.db.ReadAll(
		ctx,
		filters,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.Appointments),
	)
	if err != nil {
		s.logger.Error("Failed to list appointments", zap.Error(err))
		return nil, errors.New("failed to retrieve appointments from database")
	}

	// Check if we got results
	appointmentsArray, ok := results.([]map[string]interface{})
	if !ok {
		// If no appointments found, return empty array
		if results == nil {
			return []models.AppointmentResponse{}, nil
		}

		s.logger.Error("Failed to convert appointments to array", zap.Any("results", results))
		return nil, errors.New("invalid appointments data format")
	}

	// Convert to response objects
	response := make([]models.AppointmentResponse, 0, len(appointmentsArray))
	for _, appointmentMap := range appointmentsArray {
		appointment, err := s.mapToAppointmentResponse(appointmentMap)
		if err != nil {
			s.logger.Warn("Failed to convert appointment", zap.Error(err))
			continue
		}
		response = append(response, *appointment)
	}

	return response, nil
}

// GetDoctorAvailability returns available time slots for a doctor on a specific date
func (s *receptionService) GetDoctorAvailability(ctx context.Context, doctorID string, date time.Time, tenantID string) ([]map[string]interface{}, error) {
	// This is a simplified implementation
	// A real implementation would:
	// 1. Get the doctor's working hours for the given day
	// 2. Get all appointments for the doctor on that day
	// 3. Calculate available slots based on working hours and existing appointments

	// For simplicity, we'll assume doctors work from 9 AM to 5 PM
	// and each appointment slot is 30 minutes

	// Set date to beginning of day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	// Prepare filter to find all appointments for this doctor on this day
	filter := bson.M{
		"doctor_id":      doctorID,
		"tenant_id":      tenantID,
		"scheduled_time": bson.M{"$gte": startOfDay, "$lt": endOfDay},
		"status":         bson.M{"$nin": []string{string(models.AppointmentStatusCancelled)}},
	}

	// Get existing appointments
	results, err := s.db.ReadAll(
		ctx,
		filter,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.Appointments),
	)
	if err != nil {
		s.logger.Error("Failed to retrieve doctor appointments", zap.Error(err))
		return nil, errors.New("failed to check doctor availability")
	}

	// Create a map of occupied time slots
	occupiedSlots := make(map[string]bool)

	if results != nil {
		appointmentsArray, ok := results.([]map[string]interface{})
		if ok {
			for _, appt := range appointmentsArray {
				scheduledTime, ok := appt["scheduled_time"].(time.Time)
				if !ok {
					continue
				}

				duration, ok := appt["duration"].(int)
				if !ok {
					duration = 30 // Default duration if not specified
				}

				// Mark all 30-minute slots covered by this appointment as occupied
				for i := 0; i < duration; i += 30 {
					slotTime := scheduledTime.Add(time.Duration(i) * time.Minute)
					occupiedSlots[slotTime.Format(time.RFC3339)] = true
				}
			}
		}
	}

	// Generate all possible slots from 9 AM to 5 PM
	availableSlots := []map[string]interface{}{}
	workStart := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, date.Location())
	workEnd := time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, date.Location())

	// Use 30-minute intervals
	for slotStart := workStart; slotStart.Before(workEnd); slotStart = slotStart.Add(30 * time.Minute) {
		slotEnd := slotStart.Add(30 * time.Minute)
		slotKey := slotStart.Format(time.RFC3339)

		slot := map[string]interface{}{
			"start_time":   slotStart,
			"end_time":     slotEnd,
			"is_available": !occupiedSlots[slotKey],
		}

		availableSlots = append(availableSlots, slot)
	}

	return availableSlots, nil
}

// Helper method to convert a database map to an AppointmentResponse
func (s *receptionService) mapToAppointmentResponse(appointmentMap map[string]interface{}) (*models.AppointmentResponse, error) {
	// Extract required fields
	appointmentID, ok := appointmentMap["_id"].(string)
	if !ok {
		return nil, errors.New("invalid appointment ID format")
	}

	patientID, ok := appointmentMap["patient_id"].(string)
	if !ok {
		return nil, errors.New("invalid patient ID format")
	}

	patientName, ok := appointmentMap["patient_name"].(string)
	if !ok {
		return nil, errors.New("invalid patient name format")
	}

	doctorID, ok := appointmentMap["doctor_id"].(string)
	if !ok {
		return nil, errors.New("invalid doctor ID format")
	}

	doctorName, ok := appointmentMap["doctor_name"].(string)
	if !ok {
		return nil, errors.New("invalid doctor name format")
	}

	scheduledTime, ok := appointmentMap["scheduled_time"].(time.Time)
	if !ok {
		return nil, errors.New("invalid scheduled time format")
	}

	duration, ok := appointmentMap["duration"].(int)
	if !ok {
		durationFloat, okFloat := appointmentMap["duration"].(float64)
		if okFloat {
			duration = int(durationFloat)
		} else {
			return nil, errors.New("invalid duration format")
		}
	}

	appointmentTypeStr, ok := appointmentMap["appointment_type"].(string)
	if !ok {
		return nil, errors.New("invalid appointment type format")
	}
	appointmentType := models.AppointmentType(appointmentTypeStr)

	statusStr, ok := appointmentMap["status"].(string)
	if !ok {
		return nil, errors.New("invalid status format")
	}
	status := models.AppointmentStatus(statusStr)

	notes, _ := appointmentMap["notes"].(string)

	createdAt, ok := appointmentMap["created_at"].(time.Time)
	if !ok {
		createdAt = time.Now() // Fallback to current time if not available
	}

	// Create response object
	response := &models.AppointmentResponse{
		AppointmentID: appointmentID,
		PatientID:     patientID,
		PatientName:   patientName,
		DoctorID:      doctorID,
		DoctorName:    doctorName,
		ScheduledTime: scheduledTime,
		Duration:      duration,
		Type:          appointmentType,
		Status:        status,
		Notes:         notes,
		CreatedAt:     createdAt,
	}

	return response, nil
}

// Helper method to check if a doctor is available at a specific time
func (s *receptionService) checkDoctorAvailability(ctx context.Context, doctorID string, scheduledTime time.Time, duration int, tenantID string) (bool, error) {
	// Check if the time falls within working hours
	// Assuming working hours are 9 AM to 5 PM on all days
	hour := scheduledTime.Hour()
	if hour < 9 || hour >= 17 {
		return false, nil
	}

	// Calculate the end time of the appointment
	endTime := scheduledTime.Add(time.Duration(duration) * time.Minute)

	// Prepare filter to find conflicting appointments
	filter := bson.M{
		"doctor_id": doctorID,
		"tenant_id": tenantID,
		"status":    bson.M{"$nin": []string{string(models.AppointmentStatusCancelled)}},
		"$or": []bson.M{
			{
				// Appointment starts during another appointment
				"scheduled_time": bson.M{
					"$lt":  endTime,
					"$gte": scheduledTime,
				},
			},
			{
				// Another appointment starts during this appointment
				"scheduled_time": bson.M{
					"$lt": scheduledTime,
				},
				"$expr": bson.M{
					"$gte": []interface{}{
						bson.M{"$add": []interface{}{
							"$scheduled_time",
							bson.M{"$multiply": []interface{}{"$duration", 60 * 1000}},
						}},
						scheduledTime,
					},
				},
			},
		},
	}

	// Query database for conflicting appointments
	results, err := s.db.ReadAll(
		ctx,
		filter,
		db.WithDatabaseName(config.DatabaseNames.CoreDB),
		db.WithCollectionName(config.CollectionNames.Appointments),
	)
	if err != nil {
		s.logger.Error("Failed to check doctor availability", zap.Error(err))
		return false, errors.New("failed to check doctor availability")
	}

	// If any conflicting appointments found, the doctor is not available
	if results != nil {
		appointmentsArray, ok := results.([]map[string]interface{})
		if ok && len(appointmentsArray) > 0 {
			return false, nil
		}
	}

	return true, nil
}
