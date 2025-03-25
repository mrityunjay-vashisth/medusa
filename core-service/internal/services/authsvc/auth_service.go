package authsvc

import (
	"context"
	"crypto/rand"
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
	Register(ctx context.Context, req models.AuthRegisterRequest) (*authpb.RegisterUserResponse, error)
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

func (a *authService) Register(ctx context.Context, req models.AuthRegisterRequest) (*authpb.RegisterUserResponse, error) {
	// Call the auth service via gRPC
	var err error
	if req.Password == "" {
		req.Password = generateRandomPassword()
	}
	if err != nil {
		a.Logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}
	client := a.GetClient()
	resp, err := client.RegisterUser(ctx, &authpb.RegisterUserRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Role:     req.Role,
		TenantId: req.TenantId,
	})

	if err != nil {
		a.Logger.Error("Failed to register user", zap.Error(err))
		return nil, err
	}

	a.Logger.Info("Credentials prepared for",
		zap.String("email", req.Email),
		zap.String("username", req.Username),
		zap.String("role", req.Role),
		zap.String("password", req.Password),
	)

	return resp, nil
}

// GetClient returns the gRPC AuthServiceClient
func (a *authService) GetClient() authpb.AuthServiceClient {
	return a.client
}

func generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
