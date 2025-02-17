package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("my-secret-key")

type authService struct {
	authpb.UnimplementedAuthServiceServer
	client *mongo.Client
}

type user struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
	Role     string `bson:"role"`
	Email    string `bson:"email"`
}

type claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(client *mongo.Client) *authService {
	return &authService{client: client}
}

func (s *authService) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return nil, errors.New("username, password and email are required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	collection := s.client.Database("authdb").Collection("users")
	existingUser := collection.FindOne(ctx, bson.M{"email": req.Email})
	if existingUser.Err() == nil {
		return nil, errors.New("email already registered, try logging in")
	}
	_, err = collection.InsertOne(ctx, user{
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     "admin",
		Email:    req.Email,
	})
	if err != nil {
		return nil, err
	}

	return &authpb.RegisterResponse{Message: "User registered successfully"}, nil
}

func (s *authService) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	collection := s.client.Database("authdb").Collection("users")
	var u user
	err := collection.FindOne(ctx, bson.M{"username": req.Username}).Decode(&u)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &claims{
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return nil, err
	}
	return &authpb.LoginResponse{
		Token:   tokenString,
		Message: "Successfully logged In",
		Email:   u.Email,
	}, nil
}

func (s *authService) CheckAccess(ctx context.Context, req *authpb.CheckAccessRequest) (*authpb.CheckAccessResponse, error) {
	claims := &claims{}
	token, err := jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Role != "admin" {
		return nil, errors.New("access denied")
	}
	return &authpb.CheckAccessResponse{Message: "Access granted"}, nil

}
