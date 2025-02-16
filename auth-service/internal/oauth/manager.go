package oauth

import "errors"

// Manager holds available OAuth providers.
type Manager struct {
	Providers map[string]oAuthProvider
}

func NewManager() *Manager {
	return &Manager{Providers: make(map[string]oAuthProvider)}
}

func (m *Manager) RegisterProvider(name string, provider oAuthProvider) {
	m.Providers[name] = provider
}

func (m *Manager) GetProvider(name string) (oAuthProvider, error) {
	provider, exists := m.Providers[name]
	if !exists {
		return nil, errors.New("OAuth provider not found")
	}
	return provider, nil
}
