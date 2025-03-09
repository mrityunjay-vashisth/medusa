package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mrityunjay-vashisth/core-service/internal/apiserver"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/core-service/internal/services"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://192.168.1.14:27017"
	}

	dbConfig := db.DBConfig{
		Type:           db.MongoDB,
		URI:            mongoURI,
		DatabaseName:   "coredb",
		CollectionName: "entities",
	}

	var dbClient db.DBClientInterface = db.NewDBClient(dbConfig)
	if err := dbClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	serviceMg := services.NewServiceManager(dbClient, logger)
	apiServer := apiserver.NewAPIServer(dbClient, serviceMg.GetRegistry(), logger)

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{}, // Your React app's URL
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Session-Token"},
		AllowCredentials: true,
		// Debug mode for troubleshooting
		Debug: true,
	})

	corsHandler := corsMiddleware.Handler(apiServer.Router)
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080"
	}
	log.Println("Core API Server running on port " + apiPort + "...")
	log.Fatal(http.ListenAndServe(":"+apiPort, corsHandler))
}
