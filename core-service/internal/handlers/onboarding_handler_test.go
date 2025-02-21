package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/handlers"
	"github.com/stretchr/testify/assert"
)

var (
	jwtKey = []byte("my-secret-key")
)

// MockDBClient implements DBClientInterface for testing
type MockDBClient struct {
	ConnectFn func(ctx context.Context) error
	CreateFn  func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error)
	ReadFn    func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error)
	ReadAllFn func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error)
	DeleteFn  func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error)
}

// MockClaims defines JWT claims for testing purposes
type MockClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// Connect mock implementation (does nothing for tests)
func (m *MockDBClient) Connect(ctx context.Context) error {
	if m.ConnectFn != nil {
		return m.ConnectFn(ctx)
	}
	return nil
}

// Create mock implementation
func (m *MockDBClient) Create(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, data, opts...)
	}
	return nil, errors.New("CreateFn not implemented")
}

// Read mock implementation
func (m *MockDBClient) Read(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
	if m.ReadFn != nil {
		return m.ReadFn(ctx, filter, opts...)
	}
	return nil, errors.New("ReadFn not implemented")
}

// ReadAll mock implementation
func (m *MockDBClient) ReadAll(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
	if m.ReadAllFn != nil {
		// Ensure we return it as `interface{}`
		data, err := m.ReadAllFn(ctx, filter, opts...)
		return data, err
	}
	return nil, errors.New("ReadAllFn not implemented")
}

// Delete mock implementation
func (m *MockDBClient) Delete(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, filter, opts...)
	}
	return nil, errors.New("DeleteFn not implemented")
}

// Helper function to create a JWT token for testing
func generateToken(role string) string {
	claims := &MockClaims{
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

func TestOnboardTenant_InvalidSubPath(t *testing.T) {
	mockDB := &MockDBClient{
		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return nil, nil // No existing request
		},
		CreateFn: func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return "mocked-id", nil
		},
	}

	handler := handlers.NewOnboardingHandler(mockDB)

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

// Test OnboardTenant (Successful Case)
func TestOnboardTenant_Success(t *testing.T) {
	mockDB := &MockDBClient{
		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return nil, nil // No existing request
		},
		CreateFn: func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return "mocked-id", nil
		},
	}

	handler := handlers.NewOnboardingHandler(mockDB)

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
	mockDB := &MockDBClient{
		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return nil, nil // No existing request
		},
		CreateFn: func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return "mocked-id", nil
		},
	}

	handler := handlers.NewOnboardingHandler(mockDB)

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

func TestOnboardTenant_BadBody(t *testing.T) {
	mockDB := &MockDBClient{
		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return nil, nil // No existing request
		},
		CreateFn: func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return "mocked-id", nil
		},
	}

	handler := handlers.NewOnboardingHandler(mockDB)

	reqBody := `{"organization_n": "Test Corp", "email": "test@example.com", "role": "admin"}`
	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"subpath": "onboard"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Contains(t, resp["message"], "Invalid request body")
}

func TestOnboardTenant_BadBodyJSON(t *testing.T) {
	mockDB := &MockDBClient{
		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return nil, nil // No existing request
		},
		CreateFn: func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return "mocked-id", nil
		},
	}

	handler := handlers.NewOnboardingHandler(mockDB)

	reqBody := `{"organization_" "Test Corp", "email": "test@example.com", "role": "admin"}`
	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"subpath": "onboard"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Contains(t, resp["message"], "Invalid request body")
}

func TestOnboardTenant_CreateFailure(t *testing.T) {
	mockDB := &MockDBClient{
		CreateFn: func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return nil, errors.New("mocked create error")
		},
		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return nil, nil // No existing request
		},
	}

	handler := handlers.NewOnboardingHandler(mockDB)

	reqBody := `{"organization_name": "Test Corp", "email": "test@example.com", "role": "admin"}`
	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"subpath": "onboard"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	assert.Contains(t, resp["message"], "Failed to onboard tenant")
}

// Test OnboardTenant (Existing Request Conflict)
func TestOnboardTenant_Conflict(t *testing.T) {
	mockDB := &MockDBClient{
		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return map[string]interface{}{"email": "test@example.com"}, nil // Existing request found
		},
	}

	handler := handlers.NewOnboardingHandler(mockDB)

	reqBody := `{"organization_name": "Test Corp", "email": "test@example.com", "role": "admin"}`
	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	req = mux.SetURLVars(req, map[string]string{"subpath": "onboard"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

// Test GetPendingRequests (Success)
func TestGetPendingRequests_Success(t *testing.T) {
	mockDB := &MockDBClient{
		ReadAllFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return []map[string]interface{}{
				{"email": "test@example.com", "status": "pending"},
			}, nil
		},
	}

	handler := handlers.NewOnboardingHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/onboarding", nil)
	req.Header.Set("Authorization", generateToken("superuser"))
	req = mux.SetURLVars(req, map[string]string{"subpath": "pending"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var requests []map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&requests)
	assert.Len(t, requests, 1)
}

// Test ApproveOnboarding (Success)
func TestApproveOnboarding_Success(t *testing.T) {
	mockDB := &MockDBClient{
		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return map[string]interface{}{"email": "test@example.com", "status": "pending"}, nil
		},
		CreateFn: func(ctx context.Context, data map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return "mocked-id", nil
		},
		DeleteFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return 1, nil
		},
	}

	handler := handlers.NewOnboardingHandler(mockDB)

	reqBody := `{"request_id": "12345"}`
	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
	req.Header.Set("Authorization", generateToken("superuser"))
	req.Header.Set("Content-Type", "application/json")

	req = mux.SetURLVars(req, map[string]string{"subpath": "approve"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Contains(t, resp["message"], "Onboarding request approved")
}

// Test ApproveOnboarding (Invalid Request ID)
func TestApproveOnboarding_InvalidRequest(t *testing.T) {
	mockDB := &MockDBClient{
		ReadFn: func(ctx context.Context, filter map[string]interface{}, opts ...db.DBOption) (interface{}, error) {
			return nil, errors.New("not found")
		},
	}

	handler := handlers.NewOnboardingHandler(mockDB)

	reqBody := `{"request_id": "invalid_id"}`
	req := httptest.NewRequest(http.MethodPost, "/onboarding", bytes.NewBufferString(reqBody))
	req.Header.Set("Authorization", generateToken("superuser"))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"subpath": "approve"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Test Unauthorized Access
func TestUnauthorizedAccess(t *testing.T) {
	mockDB := &MockDBClient{}
	handler := handlers.NewOnboardingHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/onboarding", nil)
	req.Header.Set("Authorization", generateToken("user")) // Not superuser
	req = mux.SetURLVars(req, map[string]string{"subpath": "pending"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}
