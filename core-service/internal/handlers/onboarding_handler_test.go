package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/handlers"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"github.com/mrityunjay-vashisth/core-service/test/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	jwtKey = []byte("my-secret-key")
)

// Helper function to create a JWT token for testing
func generateToken(role string) string {
	claims := &mocks.MockClaims{
		Username: "testuser",
		Role:     role,
		TenantID: uuid.New().String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtKey)
	return tokenString
}

func TestOnboardTenant_Success(t *testing.T) {
	mockService := &mocks.MockOnboardingService{
		OnboardTenantFunc: func(ctx context.Context, req models.OnboardingRequest) (string, error) {
			return "12345", nil
		},
	}

	logger := zap.NewNop() // Use no-op logger for testing
	handler := handlers.NewOnboardingHandler(mockService, logger)

	requestPayload := `{"organization_name": "TestOrg", "email": "test@example.com", "role": "admin"}`
	req := httptest.NewRequest(http.MethodPost, "/onboarding/onboard", bytes.NewBuffer([]byte(requestPayload)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.OnboardTenant(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]string
	json.NewDecoder(resp.Body).Decode(&response)
	if response["request_id"] != "12345" {
		t.Errorf("Expected request_id 12345, got %s", response["request_id"])
	}
}

func TestOnboardTenant_InvalidBody(t *testing.T) {
	mockService := &mocks.MockOnboardingService{}
	logger := zap.NewNop()
	handler := handlers.NewOnboardingHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodPost, "/onboarding/onboard", bytes.NewBuffer([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.OnboardTenant(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestOnboardTenant_InvalidSubPath(t *testing.T) {
	mockService := &mocks.MockOnboardingService{
		OnboardTenantFunc: func(ctx context.Context, req models.OnboardingRequest) (string, error) {
			return "12345", nil
		},
	}

	logger := zap.NewNop() // Use no-op logger for testing
	handler := handlers.NewOnboardingHandler(mockService, logger)

	reqBody := `{"organization_name": "Test Corp", "email": "test@example.com", "role": "admin"}`
	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
	req.Header.Set("Authorization", generateToken("superuser"))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"subpath": "ind"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Contains(t, resp["message"], "Invalid Onboarding API")
}

func TestGetPendingRequests_Success(t *testing.T) {
	mockService := &mocks.MockOnboardingService{
		GetPendingRequestsFunc: func(ctx context.Context) (interface{}, error) {
			return []models.EntityMetadata{
				{RequestID: "12345", Status: "pending"},
			}, nil
		},
	}

	logger := zap.NewNop()
	handler := handlers.NewOnboardingHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodGet, "/onboarding/pending", nil)
	req.Header.Set("Authorization", generateToken("superuser"))
	req = mux.SetURLVars(req, map[string]string{"subpath": "pending"})

	w := httptest.NewRecorder()
	handler.GetPendingRequests(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response []models.EntityMetadata
	json.NewDecoder(resp.Body).Decode(&response)
	if len(response) != 1 || response[0].RequestID != "12345" {
		t.Errorf("Unexpected response: %+v", response)
	}
}

func TestApproveOnboarding_Success(t *testing.T) {
	mockService := &mocks.MockOnboardingService{
		ApproveOnboardingFunc: func(ctx context.Context, requestID string) error {
			return nil
		},
	}

	logger := zap.NewNop()
	handler := handlers.NewOnboardingHandler(mockService, logger)

	requestPayload := `{"request_id": "12345"}`
	req := httptest.NewRequest(http.MethodPost, "/onboarding/approve", bytes.NewBuffer([]byte(requestPayload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generateToken("superuser"))

	w := httptest.NewRecorder()
	handler.ApproveOnboarding(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestApproveOnboarding_InvalidBody(t *testing.T) {
	mockService := &mocks.MockOnboardingService{}
	logger := zap.NewNop()
	handler := handlers.NewOnboardingHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodPost, "/onboarding/approve", bytes.NewBuffer([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generateToken("superuser"))

	w := httptest.NewRecorder()
	handler.ApproveOnboarding(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

// Test OnboardTenant (Successful Case)
func TestOnboardTenant_SuccessSubpath(t *testing.T) {
	mockService := &mocks.MockOnboardingService{
		OnboardTenantFunc: func(ctx context.Context, req models.OnboardingRequest) (string, error) {
			return "12345", nil
		},
	}
	logger := zap.NewNop()
	handler := handlers.NewOnboardingHandler(mockService, logger)

	reqBody := `{"organization_name": "Test Corp", "email": "test@example.com", "role": "admin"}`
	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"subpath": "onboard"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Contains(t, resp["message"], "Onboarding request submitted")
}

func TestOnboardTenant_BadMethod(t *testing.T) {
	mockService := &mocks.MockOnboardingService{
		OnboardTenantFunc: func(ctx context.Context, req models.OnboardingRequest) (string, error) {
			return "12345", nil
		},
	}
	logger := zap.NewNop()
	handler := handlers.NewOnboardingHandler(mockService, logger)

	reqBody := `{"organization_name": "Test Corp", "email": "test@example.com", "role": "admin"}`
	req := httptest.NewRequest(http.MethodGet, "/onboarding", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"subpath": "onboard"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Contains(t, resp["message"], "Invalid request method")
}

// func TestOnboardTenant_BadReq(t *testing.T) {
// 	mockService := &mocks.MockOnboardingService{
// 		OnboardTenantFunc: func(ctx context.Context, req models.OnboardingRequest) (string, error) {
// 			return "12345", nil
// 		},
// 	}
// 	logger := zap.NewNop()
// 	handler := handlers.NewOnboardingHandler(mockService, logger)

// 	reqBody := `{"organization_n": "Test Corp", "email": "test@example.com", "role": "admin"}`
// 	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
// 	req.Header.Set("Content-Type", "application/json")
// 	req = mux.SetURLVars(req, map[string]string{"subpath": "onboard"})

// 	rec := httptest.NewRecorder()
// 	handler.ServeHTTP(rec, req)

// 	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
// 	var resp map[string]string
// 	json.NewDecoder(rec.Body).Decode(&resp)
// 	assert.Contains(t, resp["message"], "Invalid request method")
// }

// func TestOnboardTenant_BadBody(t *testing.T) {
// 	mockDB := &MockDBClient{
// 		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return nil, nil // No existing request
// 		},
// 		CreateFn: func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return "mocked-id", nil
// 		},
// 	}

// 	handler := handlers.NewOnboardingHandler(mockDB)

// 	reqBody := `{"organization_n": "Test Corp", "email": "test@example.com", "role": "admin"}`
// 	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
// 	req.Header.Set("Content-Type", "application/json")
// 	req = mux.SetURLVars(req, map[string]string{"subpath": "onboard"})

// 	rec := httptest.NewRecorder()
// 	handler.ServeHTTP(rec, req)

// 	assert.Equal(t, http.StatusBadRequest, rec.Code)
// 	var resp map[string]string
// 	json.NewDecoder(rec.Body).Decode(&resp)
// 	assert.Contains(t, resp["message"], "Invalid request body")
// }

// func TestOnboardTenant_BadBodyJSON(t *testing.T) {
// 	mockDB := &MockDBClient{
// 		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return nil, nil // No existing request
// 		},
// 		CreateFn: func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return "mocked-id", nil
// 		},
// 	}

// 	handler := handlers.NewOnboardingHandler(mockDB)

// 	reqBody := `{"organization_" "Test Corp", "email": "test@example.com", "role": "admin"}`
// 	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
// 	req.Header.Set("Content-Type", "application/json")
// 	req = mux.SetURLVars(req, map[string]string{"subpath": "onboard"})

// 	rec := httptest.NewRecorder()
// 	handler.ServeHTTP(rec, req)

// 	assert.Equal(t, http.StatusBadRequest, rec.Code)
// 	var resp map[string]string
// 	json.NewDecoder(rec.Body).Decode(&resp)
// 	assert.Contains(t, resp["message"], "Invalid request body")
// }

// func TestOnboardTenant_CreateFailure(t *testing.T) {
// 	mockDB := &MockDBClient{
// 		CreateFn: func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return nil, errors.New("mocked create error")
// 		},
// 		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return nil, nil // No existing request
// 		},
// 	}

// 	handler := handlers.NewOnboardingHandler(mockDB)

// 	reqBody := `{"organization_name": "Test Corp", "email": "test@example.com", "role": "admin"}`
// 	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
// 	req.Header.Set("Content-Type", "application/json")
// 	req = mux.SetURLVars(req, map[string]string{"subpath": "onboard"})

// 	rec := httptest.NewRecorder()
// 	handler.ServeHTTP(rec, req)

// 	assert.Equal(t, http.StatusInternalServerError, rec.Code)
// 	var resp map[string]string
// 	err := json.NewDecoder(rec.Body).Decode(&resp)
// 	if err != nil {
// 		t.Fatalf("Failed to decode response: %v", err)
// 	}
// 	assert.Contains(t, resp["message"], "Failed to onboard tenant")
// }

// // Test OnboardTenant (Existing Request Conflict)
// func TestOnboardTenant_Conflict(t *testing.T) {
// 	mockDB := &MockDBClient{
// 		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return map[string]interface{}{"email": "test@example.com"}, nil // Existing request found
// 		},
// 	}

// 	handler := handlers.NewOnboardingHandler(mockDB)

// 	reqBody := `{"organization_name": "Test Corp", "email": "test@example.com", "role": "admin"}`
// 	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
// 	req.Header.Set("Content-Type", "application/json")

// 	req = mux.SetURLVars(req, map[string]string{"subpath": "onboard"})
// 	rec := httptest.NewRecorder()
// 	handler.ServeHTTP(rec, req)
// 	var resp map[string]string
// 	json.NewDecoder(rec.Body).Decode(&resp)
// 	assert.Equal(t, http.StatusConflict, rec.Code)
// }

// // Test GetPendingRequests (Success)
// func TestGetPendingRequests_Success(t *testing.T) {
// 	mockDB := &MockDBClient{
// 		ReadAllFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return []map[string]interface{}{
// 				{"email": "test@example.com", "status": "pending"},
// 			}, nil
// 		},
// 	}

// 	handler := handlers.NewOnboardingHandler(mockDB)

// 	req := httptest.NewRequest(http.MethodGet, "/onboarding", nil)
// 	req.Header.Set("Authorization", generateToken("superuser"))
// 	req = mux.SetURLVars(req, map[string]string{"subpath": "pending"})

// 	rec := httptest.NewRecorder()
// 	handler.ServeHTTP(rec, req)
// 	assert.Equal(t, http.StatusOK, rec.Code)

// 	var requests []map[string]interface{}
// 	json.NewDecoder(rec.Body).Decode(&requests)
// 	assert.Len(t, requests, 1)
// }

// // Test ApproveOnboarding (Success)
// func TestApproveOnboarding_Success(t *testing.T) {
// 	mockDB := &MockDBClient{
// 		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return map[string]interface{}{"email": "test@example.com", "status": "pending"}, nil
// 		},
// 		CreateFn: func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return "mocked-id", nil
// 		},
// 		DeleteFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return 1, nil
// 		},
// 	}

// 	handler := handlers.NewOnboardingHandler(mockDB)

// 	reqBody := `{"request_id": "12345"}`
// 	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
// 	req.Header.Set("Authorization", generateToken("superuser"))
// 	req.Header.Set("Content-Type", "application/json")

// 	req = mux.SetURLVars(req, map[string]string{"subpath": "approve"})

// 	rec := httptest.NewRecorder()
// 	handler.ServeHTTP(rec, req)
// 	assert.Equal(t, http.StatusOK, rec.Code)
// 	var resp map[string]string
// 	json.NewDecoder(rec.Body).Decode(&resp)
// 	assert.Contains(t, resp["message"], "Onboarding request approved")
// }

// // Test ApproveOnboarding (Invalid Request ID)
// func TestApproveOnboarding_InvalidRequest(t *testing.T) {
// 	mockDB := &MockDBClient{
// 		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
// 			return nil, errors.New("not found")
// 		},
// 	}

// 	handler := handlers.NewOnboardingHandler(mockDB)

// 	reqBody := `{"request_id": "invalid_id"}`
// 	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
// 	req.Header.Set("Authorization", generateToken("superuser"))
// 	req.Header.Set("Content-Type", "application/json")
// 	req = mux.SetURLVars(req, map[string]string{"subpath": "approve"})

// 	rec := httptest.NewRecorder()
// 	handler.ServeHTTP(rec, req)
// 	assert.Equal(t, http.StatusBadRequest, rec.Code)
// }

// // Test Unauthorized Access
// func TestUnauthorizedAccess(t *testing.T) {
// 	mockDB := &MockDBClient{}
// 	handler := handlers.NewOnboardingHandler(mockDB)

// 	req := httptest.NewRequest(http.MethodGet, "/onboarding", nil)
// 	req.Header.Set("Authorization", generateToken("user")) // Not superuser
// 	req = mux.SetURLVars(req, map[string]string{"subpath": "pending"})

// 	rec := httptest.NewRecorder()
// 	handler.ServeHTTP(rec, req)

// 	assert.Equal(t, http.StatusForbidden, rec.Code)
// }
