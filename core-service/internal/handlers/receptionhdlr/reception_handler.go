package receptionhdlr

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/handlers/utility"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"github.com/mrityunjay-vashisth/core-service/internal/registry"
	"github.com/mrityunjay-vashisth/core-service/internal/services/receptionsvc"
	"go.uber.org/zap"
)

type ReceptionHandlerInterface interface {
	ListAppointments(w http.ResponseWriter, r *http.Request)
	CreateAppointment(w http.ResponseWriter, r *http.Request)
	GetAppointmentByID(w http.ResponseWriter, r *http.Request)
	UpdateAppointment(w http.ResponseWriter, r *http.Request)
	CancelAppointment(w http.ResponseWriter, r *http.Request)
	GetDoctorAvailability(w http.ResponseWriter, r *http.Request)
}

type receptionHandler struct {
	registry registry.ServiceRegistry
	logger   *zap.Logger
}

func NewReceptionHandler(registry registry.ServiceRegistry, logger *zap.Logger) ReceptionHandlerInterface {
	return &receptionHandler{
		registry: registry,
		logger:   logger,
	}
}

// getReceptionService retrieves the reception service from the registry
func (h *receptionHandler) getReceptionService() (receptionsvc.Service, error) {
	service, ok := h.registry.Get(registry.ReceptionService).(receptionsvc.Service)
	if !ok {
		h.logger.Error("Failed to get reception service from registry")
		return nil, errors.New("internal service error")
	}
	return service, nil
}

// extractTenantIDFromContext extracts the tenant ID from the request context
func (h *receptionHandler) extractTenantIDFromContext(r *http.Request) (string, error) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		return "", errors.New("tenant ID not found in context")
	}
	return tenantID, nil
}

// extractUsernameFromContext extracts the username from the request context
func (h *receptionHandler) extractUsernameFromContext(r *http.Request) (string, error) {
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		return "", errors.New("username not found in context")
	}
	return username, nil
}

// ListAppointments handles requests to list appointments with optional filtering
func (h *receptionHandler) ListAppointments(w http.ResponseWriter, r *http.Request) {
	// Extract tenant ID from context
	tenantID, err := h.extractTenantIDFromContext(r)
	if err != nil {
		utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	// Get reception service
	service, err := h.getReceptionService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Extract query parameters
	query := r.URL.Query()
	filters := make(map[string]interface{})

	// Add filters if present in query
	if doctorID := query.Get("doctor_id"); doctorID != "" {
		filters["doctor_id"] = doctorID
	}

	if patientID := query.Get("patient_id"); patientID != "" {
		filters["patient_id"] = patientID
	}

	if dateFrom := query.Get("date_from"); dateFrom != "" {
		filters["date_from"] = dateFrom
	}

	if dateTo := query.Get("date_to"); dateTo != "" {
		filters["date_to"] = dateTo
	}

	if status := query.Get("status"); status != "" {
		filters["status"] = status
	}

	// Call service to list appointments
	appointments, err := service.ListAppointments(r.Context(), filters, tenantID)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Failed to list appointments: "+err.Error())
		return
	}

	// Respond with appointments
	utility.RespondWithJSON(w, http.StatusOK, appointments)
}

// CreateAppointment handles requests to create a new appointment
func (h *receptionHandler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	// Extract tenant ID from context
	tenantID, err := h.extractTenantIDFromContext(r)
	if err != nil {
		utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	// Extract username (creator) from context
	username, err := h.extractUsernameFromContext(r)
	if err != nil {
		utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	// Parse request body
	var req models.AppointmentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.PatientID == "" || req.PatientName == "" || req.DoctorID == "" || req.DoctorName == "" || req.Duration <= 0 {
		utility.RespondWithError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// Get reception service
	service, err := h.getReceptionService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Call service to create appointment
	appointment, err := service.CreateAppointment(r.Context(), req, tenantID, username)
	if err != nil {
		if err.Error() == "doctor is not available at the requested time" {
			utility.RespondWithError(w, http.StatusConflict, err.Error())
		} else {
			utility.RespondWithError(w, http.StatusInternalServerError, "Failed to create appointment: "+err.Error())
		}
		return
	}

	// Respond with created appointment
	utility.RespondWithJSON(w, http.StatusCreated, appointment)
}

// GetAppointmentByID handles requests to retrieve a specific appointment
func (h *receptionHandler) GetAppointmentByID(w http.ResponseWriter, r *http.Request) {
	// Extract tenant ID from context
	tenantID, err := h.extractTenantIDFromContext(r)
	if err != nil {
		utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	// Extract appointment ID from URL path
	vars := mux.Vars(r)
	appointmentID := vars["id"]
	if appointmentID == "" {
		utility.RespondWithError(w, http.StatusBadRequest, "Missing appointment ID")
		return
	}

	// Get reception service
	service, err := h.getReceptionService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Call service to get appointment
	appointment, err := service.GetAppointmentByID(r.Context(), appointmentID, tenantID)
	if err != nil {
		if err.Error() == "appointment not found" {
			utility.RespondWithError(w, http.StatusNotFound, err.Error())
		} else {
			utility.RespondWithError(w, http.StatusInternalServerError, "Failed to get appointment: "+err.Error())
		}
		return
	}

	// Respond with appointment
	utility.RespondWithJSON(w, http.StatusOK, appointment)
}

// UpdateAppointment handles requests to update an existing appointment
func (h *receptionHandler) UpdateAppointment(w http.ResponseWriter, r *http.Request) {
	// Extract tenant ID from context
	tenantID, err := h.extractTenantIDFromContext(r)
	if err != nil {
		utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	// Extract appointment ID from URL path
	vars := mux.Vars(r)
	appointmentID := vars["id"]
	if appointmentID == "" {
		utility.RespondWithError(w, http.StatusBadRequest, "Missing appointment ID")
		return
	}

	// Parse request body
	var req models.AppointmentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Get reception service
	service, err := h.getReceptionService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Call service to update appointment
	appointment, err := service.UpdateAppointment(r.Context(), appointmentID, req, tenantID)
	if err != nil {
		if err.Error() == "appointment not found" {
			utility.RespondWithError(w, http.StatusNotFound, err.Error())
		} else if err.Error() == "doctor is not available at the requested time" {
			utility.RespondWithError(w, http.StatusConflict, err.Error())
		} else {
			utility.RespondWithError(w, http.StatusInternalServerError, "Failed to update appointment: "+err.Error())
		}
		return
	}

	// Respond with updated appointment
	utility.RespondWithJSON(w, http.StatusOK, appointment)
}

// CancelAppointment handles requests to cancel an appointment
func (h *receptionHandler) CancelAppointment(w http.ResponseWriter, r *http.Request) {
	// Extract tenant ID from context
	tenantID, err := h.extractTenantIDFromContext(r)
	if err != nil {
		utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	// Extract appointment ID from URL path
	vars := mux.Vars(r)
	appointmentID := vars["id"]
	if appointmentID == "" {
		utility.RespondWithError(w, http.StatusBadRequest, "Missing appointment ID")
		return
	}

	// Parse request body to get cancellation reason
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.Reason == "" {
		utility.RespondWithError(w, http.StatusBadRequest, "Cancellation reason is required")
		return
	}

	// Get reception service
	service, err := h.getReceptionService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Call service to cancel appointment
	appointment, err := service.CancelAppointment(r.Context(), appointmentID, req.Reason, tenantID)
	if err != nil {
		if err.Error() == "appointment not found" {
			utility.RespondWithError(w, http.StatusNotFound, err.Error())
		} else {
			utility.RespondWithError(w, http.StatusInternalServerError, "Failed to cancel appointment: "+err.Error())
		}
		return
	}

	// Respond with cancelled appointment
	utility.RespondWithJSON(w, http.StatusOK, appointment)
}

// GetDoctorAvailability handles requests to check a doctor's availability
func (h *receptionHandler) GetDoctorAvailability(w http.ResponseWriter, r *http.Request) {
	// Extract tenant ID from context
	tenantID, err := h.extractTenantIDFromContext(r)
	if err != nil {
		utility.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	// Extract query parameters
	query := r.URL.Query()
	doctorID := query.Get("doctor_id")
	dateStr := query.Get("date")

	if doctorID == "" {
		utility.RespondWithError(w, http.StatusBadRequest, "Doctor ID is required")
		return
	}

	if dateStr == "" {
		utility.RespondWithError(w, http.StatusBadRequest, "Date is required")
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		utility.RespondWithError(w, http.StatusBadRequest, "Invalid date format. Expected YYYY-MM-DD")
		return
	}

	// Get reception service
	service, err := h.getReceptionService()
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Call service to get availability
	availability, err := service.GetDoctorAvailability(r.Context(), doctorID, date, tenantID)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Failed to get doctor availability: "+err.Error())
		return
	}

	// Respond with availability
	utility.RespondWithJSON(w, http.StatusOK, availability)
}
