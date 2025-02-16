// package oauth

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"

// 	"golang.org/x/oauth2"
// 	"golang.org/x/oauth2/google"
// )

// type googleProvider struct {
// 	config *oauth2.Config
// }

// func NewGoogleProvider(cliendID, clientSecret string) *googleProvider {
// 	return &googleProvider{
// 		config: &oauth2.Config{
// 			ClientID:     cliendID,
// 			ClientSecret: clientSecret,
// 			RedirectURL:  "http://192.168.1.12:8080/auth/google/callback",
// 			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
// 			Endpoint:     google.Endpoint,
// 		},
// 	}
// }

// func (g *googleProvider) GetConfig() *oauth2.Config {
// 	return g.config
// }

// func (g *googleProvider) GetUserInfo(token *oauth2.Token) (map[string]interface{}, error) {
// 	resp, err := http.Get(fmt.Sprintf("https://www.googleapis.com/oauth2/v2/userinfo?access_token=%s", token.AccessToken))
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var userInfo map[string]interface{}
// 	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
// 		return nil, err
// 	}

// 	return userInfo, nil
// }

package oauth

import (
	"errors"

	"golang.org/x/oauth2"
)

type MockGoogleProvider struct{}

func NewMockGoogleProvider() *MockGoogleProvider {
	return &MockGoogleProvider{}
}

// Return a fake OAuth config (not calling real Google OAuth)
func (g *MockGoogleProvider) GetConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "mock-client-id",
		ClientSecret: "mock-client-secret",
		RedirectURL:  "http://192.168.1.14:8080/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://192.168.1.14:8080/mock/oauth/google/auth",
			TokenURL: "http://192.168.1.14:8080/mock/oauth/google/token",
		},
	}
}

// Mock user authentication response
func (g *MockGoogleProvider) GetUserInfo(token *oauth2.Token) (map[string]interface{}, error) {
	if token.AccessToken != "mock-access-token" {
		return nil, errors.New("invalid access token")
	}

	return map[string]interface{}{
		"email": "mockuser@example.com",
		"name":  "Mock User",
	}, nil
}
