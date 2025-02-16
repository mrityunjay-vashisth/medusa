package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mrityunjay-vashisth/auth-service/internal/oauth"
	"github.com/mrityunjay-vashisth/auth-service/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type oAuthService struct {
	proto.UnimplementedOAuthServiceServer
	manager *oauth.Manager
	client  *mongo.Client
}

func NewOAuthService(manager *oauth.Manager, client *mongo.Client) *oAuthService {
	return &oAuthService{manager: manager, client: client}
}

func (s *oAuthService) OAuthLogin(ctx context.Context, req *proto.OAuthLoginRequest) (*proto.OAuthLoginResponse, error) {
	provider, err := s.manager.GetProvider(req.Provider)
	if err != nil {
		return nil, errors.New("unsupported provider")
	}
	url := provider.GetConfig().AuthCodeURL("state-token")
	return &proto.OAuthLoginResponse{Url: url}, nil
}

func (s *oAuthService) OAuthCallback(ctx context.Context, req *proto.OAuthCallbackRequest) (*proto.OAuthCallbackResponse, error) {
	provider, err := s.manager.GetProvider(req.Provider)
	if err != nil {
		return nil, errors.New("unsupported provider")
	}

	if req.Code == "mock-auth-code" {
		// Generate mock token
		userInfo := map[string]interface{}{
			"email": "mockuser@example.com",
			"name":  "Mock User",
		}

		// Generate JWT for mock user
		tokenString := generateJWT(userInfo["name"].(string), "user")
		return &proto.OAuthCallbackResponse{
			Token:   tokenString,
			Message: "Mock Login Success",
			Email:   userInfo["email"].(string),
		}, nil
	}

	// Exchange code for token
	token, err := provider.GetConfig().Exchange(ctx, req.Code)
	if err != nil {
		return nil, errors.New("failed to exchange code for token")
	}

	// Fetch user info from provider
	userInfo, err := provider.GetUserInfo(token)
	if err != nil {
		return nil, errors.New("failed to fetch user info")
	}

	email, ok := userInfo["email"].(string)
	if !ok || email == "" {
		return nil, errors.New("email not found in provider response")
	}

	collection := s.client.Database("authdb").Collection("users")
	var u user

	// Check if user exists
	err = collection.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		// Register new user
		u = user{
			Username: userInfo["name"].(string),
			Email:    email,
			Role:     "user",
		}
		_, err := collection.InsertOne(ctx, u)
		if err != nil {
			return nil, errors.New("failed to register user")
		}
	} else if err != nil {
		return nil, errors.New("database error")
	}

	// Generate JWT Token
	tokenString := generateJWT(u.Username, u.Role)
	return &proto.OAuthCallbackResponse{
		Token:   tokenString,
		Message: "Successfully Logged In",
		Email:   u.Email,
	}, nil
}

// Helper function to generate JWT token
func generateJWT(username, role string) string {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Println("Error generating JWT:", err)
		return ""
	}

	return tokenString
}
