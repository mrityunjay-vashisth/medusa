package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mrityunjay-vashisth/core-service/internal/auth"
	"github.com/mrityunjay-vashisth/core-service/internal/onboarding"
	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://192.168.1.12:27017"))
	if err != nil {
		log.Fatal(err)
	}

	authConn, err := grpc.Dial("192.168.1.14:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to auth-service: %v", err)
	}
	authClient := authpb.NewAuthServiceClient(authConn)

	authHandler := auth.NewAuthClient(authClient)

	httpHandlers := onboarding.NewOnboardingHandler(client)
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	router.OPTIONS("/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Status(200)
	})

	router.POST("/onboard", httpHandlers.OnboardTenant) // Public
	router.POST("/login", authHandler.Login)
	router.GET("/pending-requests", httpHandlers.GetPendingRequests) // Admin Only, Require JWT
	router.POST("/approve-requests", httpHandlers.ApproveOnboarding) // Requires JWT

	log.Println("Core Service HTTP API running on port 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}
