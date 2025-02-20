package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mrityunjay-vashisth/core-service/internal/apiserver"
	"github.com/mrityunjay-vashisth/core-service/internal/services"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbclient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://192.168.1.12:27017"))
	if err != nil {
		log.Fatal(err)
	}
	authClient := services.NewAuthClient("192.168.1.14:50051").GetClient()
	apiServer := apiserver.NewAPIServer(dbclient, authClient)

	log.Println("Core API Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", apiServer.Router))
}
