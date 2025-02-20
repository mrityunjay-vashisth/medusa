package services

import (
	"log"

	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthClient wraps gRPC AuthService
type AuthClient struct {
	client authpb.AuthServiceClient
}

// NewAuthClient initializes gRPC client connection
func NewAuthClient(authServiceAddr string) *AuthClient {
	conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to auth-service: %v", err)
	}
	return &AuthClient{client: authpb.NewAuthServiceClient(conn)}
}

// GetClient returns the gRPC AuthServiceClient
func (a *AuthClient) GetClient() authpb.AuthServiceClient {
	return a.client
}
