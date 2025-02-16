package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/mrityunjay-vashisth/auth-service/internal/auth"
	"github.com/mrityunjay-vashisth/auth-service/internal/oauth"
	"github.com/mrityunjay-vashisth/auth-service/proto"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

var client *mongo.Client

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://192.168.1.12:27017"))
	if err != nil {
		log.Fatal(err)
	}

	oauthManager := oauth.NewManager()
	//oauthManager.RegisterProvider("google", oauth.NewGoogleProvider("YOUR_GOOGLE_CLIENT_ID", "YOUR_GOOGLE_CLIENT_SECRET"))
	oauthManager.RegisterProvider("google", oauth.NewMockGoogleProvider())

	authService := auth.NewAuthService(client)
	oauthService := auth.NewOAuthService(oauthManager, client)

	listner, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterAuthServiceServer(grpcServer, authService)
	proto.RegisterOAuthServiceServer(grpcServer, oauthService)

	log.Println("gRPC server running on port 50051")
	if err := grpcServer.Serve(listner); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}
