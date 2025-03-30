package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/mrityunjay-vashisth/auth-service/internal/auth"
	"github.com/mrityunjay-vashisth/auth-service/internal/oauth"
	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

var client *mongo.Client

func setupIndexes(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a unique index on username
	collection := client.Database("authdb").Collection("users")
	// Also create an index on email for faster lookups
	emailIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	tenantIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "tenantid", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	// Create both indexes
	_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		emailIndex,
		tenantIndex,
	})

	return err
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://192.168.1.14:27017"
	}
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	jwtKeyString := os.Getenv("JWT_SECRET_KEY")
	if jwtKeyString != "" {
		auth.SetJWTKey([]byte(jwtKeyString))
	}

	if err := setupIndexes(client); err != nil {
		log.Printf("Warning: Failed to set up indexes: %v", err)
	}

	oauthManager := oauth.NewManager()
	//oauthManager.RegisterProvider("google", oauth.NewGoogleProvider("YOUR_GOOGLE_CLIENT_ID", "YOUR_GOOGLE_CLIENT_SECRET"))
	oauthManager.RegisterProvider("google", oauth.NewMockGoogleProvider())

	authService := auth.NewAuthService(client)
	oauthService := auth.NewOAuthService(oauthManager, client)

	grpcPort := os.Getenv("AUTH_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	listner, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	authpb.RegisterAuthServiceServer(grpcServer, authService)
	authpb.RegisterOAuthServiceServer(grpcServer, oauthService)

	log.Println("gRPC server running on port 50051")
	if err := grpcServer.Serve(listner); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}
