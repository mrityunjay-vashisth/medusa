package services

import (
	"context"
	"log"

	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthServices interface {
	GetClient() authpb.AuthServiceClient
	Login(context.Context, string, string) (*authpb.LoginResponse, error)
}

// AuthClient wraps gRPC AuthService
type AuthService struct {
	Logger *zap.Logger
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

func (a *AuthService) Login(ctx context.Context, username, password string) (*authpb.LoginResponse, error) {
	a.Logger.Info("Got called in auth service")
	authReq := &authpb.LoginRequest{
		Username: username,
		Password: password,
	}

	authResp, err := a.client.Login(ctx, authReq)
	return authResp, err
}

func (a *AuthService) Register(ctx context.Context) {}

// GetClient returns the gRPC AuthServiceClient
func (a *AuthService) GetClient() authpb.AuthServiceClient {
	return a.client
}
