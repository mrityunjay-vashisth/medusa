package authsvc

import (
	"context"
	"log"

	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/models"
	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Service interface {
	GetClient() authpb.AuthServiceClient
	Login(context.Context, string, string) (*authpb.LoginResponse, error)
	CreateSession(claims *models.UserClaims) (string, error)
}

// AuthClient wraps gRPC AuthService
type authService struct {
	db     db.DBClientInterface
	Logger *zap.Logger
	client authpb.AuthServiceClient
}

// NewAuthClient initializes gRPC client connection
func NewService(db db.DBClientInterface, authServiceAddr string, logger *zap.Logger) Service {
	conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to auth-service: %v", err)
	}
	return &authService{db: db, client: authpb.NewAuthServiceClient(conn), Logger: logger}
}

func (a *authService) Login(ctx context.Context, username, password string) (*authpb.LoginResponse, error) {
	a.Logger.Info("Got called in auth service")
	authReq := &authpb.LoginRequest{
		Username: username,
		Password: password,
	}

	authResp, err := a.client.Login(ctx, authReq)
	return authResp, err
}

func (a *authService) Register(ctx context.Context) {}

// GetClient returns the gRPC AuthServiceClient
func (a *authService) GetClient() authpb.AuthServiceClient {
	return a.client
}
