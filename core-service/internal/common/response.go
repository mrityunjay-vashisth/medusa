package common

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// ErrorResponse provides a standardized error response format
type ErrorResponse struct {
	Status    int    `json:"-"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// SuccessResponse provides a standardized success response format
type SuccessResponse struct {
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// RespondWithError sends a standardized error response
func RespondWithError(w http.ResponseWriter, r *http.Request, logger *zap.Logger, err error, status int, message string) {
	requestID := GetRequestIDFromContext(r.Context())

	errResponse := ErrorResponse{
		Status:    status,
		Message:   message,
		RequestID: requestID,
	}

	if err != nil {
		errResponse.Error = err.Error()
		logger.Error("HTTP error response",
			zap.Int("status", status),
			zap.String("message", message),
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("path", r.URL.Path),
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(errResponse); err != nil {
		logger.Error("Failed to encode error response", zap.Error(err))
		http.Error(w, `{"message":"Internal server error","error":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}

// RespondWithJSON sends a standardized success response
func RespondWithJSON(w http.ResponseWriter, r *http.Request, logger *zap.Logger, data interface{}, status int, message string) {
	requestID := GetRequestIDFromContext(r.Context())

	response := SuccessResponse{
		Data:      data,
		Message:   message,
		RequestID: requestID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode success response",
			zap.Error(err),
			zap.String("request_id", requestID),
		)
		http.Error(w, `{"message":"Internal server error","error":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}
