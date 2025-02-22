package mocks

import (
	"context"

	"github.com/mrityunjay-vashisth/core-service/internal/models"
)

// MockOnboardingService implements OnboardingServices for testing
type MockOnboardingService struct {
	OnboardTenantFunc      func(ctx context.Context, req models.OnboardingRequest) (string, error)
	GetPendingRequestsFunc func(ctx context.Context) (interface{}, error)
	ApproveOnboardingFunc  func(ctx context.Context, requestID string) error
}

func (m *MockOnboardingService) OnboardTenant(ctx context.Context, req models.OnboardingRequest) (string, error) {
	return m.OnboardTenantFunc(ctx, req)
}

func (m *MockOnboardingService) GetPendingRequests(ctx context.Context) (interface{}, error) {
	return m.GetPendingRequestsFunc(ctx)
}

func (m *MockOnboardingService) ApproveOnboarding(ctx context.Context, requestID string) error {
	return m.ApproveOnboardingFunc(ctx, requestID)
}
