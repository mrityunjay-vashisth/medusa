package mocks

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
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
