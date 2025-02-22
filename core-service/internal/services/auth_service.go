package services

import (
	"log"

	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthServices interface {
	GetClient() authpb.AuthServiceClient
}

// AuthClient wraps gRPC AuthService
type AuthService struct {
	client authpb.AuthServiceClient
}

// NewAuthClient initializes gRPC client connection
func NewAuthService(authServiceAddr string) *AuthService {
	conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to auth-service: %v", err)
	}
	return &AuthService{client: authpb.NewAuthServiceClient(conn)}
}

// GetClient returns the gRPC AuthServiceClient
func (a *AuthService) GetClient() authpb.AuthServiceClient {
	return a.client
}
