package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("your-secure-jwt-secret-replace-in-production")

type authService struct {
	authpb.UnimplementedAuthServiceServer
	client *mongo.Client
}

type user struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
	Role     string `bson:"role"`
	Email    string `bson:"email"`
	TenantId string `bson:"tenantid"`
}

type claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(client *mongo.Client) *authService {
	return &authService{client: client}
}

func (s *authService) RegisterUser(ctx context.Context, req *authpb.RegisterUserRequest) (*authpb.RegisterUserResponse, error) {
	// Validate input
	if req.Username == "" || req.Password == "" || req.Email == "" || req.Role == "" {
		return nil, errors.New("username, password email and role are required")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user document
	newUser := user{
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     req.Role,
		Email:    req.Email,
		TenantId: req.TenantId,
	}

	collection := s.client.Database("authdb").Collection("users")

	// Insert the user directly, relying on unique indexes
	_, err = collection.InsertOne(ctx, newUser)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("email already registered, try logging in")
		}
		return nil, errors.New("failed to register user")
	}

	return &authpb.RegisterUserResponse{Message: "User registered successfully"}, nil
}

func (s *authService) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	collection := s.client.Database("authdb").Collection("users")
	var u user
	err := collection.FindOne(ctx, bson.M{"username": req.Username}).Decode(&u)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password))
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
	log.Printf("%s", u.Role)
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

	if claims.Role != "superuser" {
		return nil, errors.New("access denied")
	}
	return &authpb.CheckAccessResponse{Message: "Access granted"}, nil

}

func SetJWTKey(key []byte) {
	jwtKey = key
}
