package services

type Container struct {
	AuthService       *AuthService
	OnboardingService *OnboardingService
	SessionService    *SessionService
}
