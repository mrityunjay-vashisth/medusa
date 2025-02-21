package handlers

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// **Storage Interface to Abstract MongoDB Calls**
type Storage interface {
	Find(ctx context.Context, collection string, filter bson.M) (*mongo.Cursor, error)
	FindOne(ctx context.Context, collection string, filter bson.M) (*mongo.SingleResult, error)
	InsertOne(ctx context.Context, collection string, document interface{}) (*mongo.InsertOneResult, error)
	DeleteOne(ctx context.Context, collection string, filter bson.M) (*mongo.DeleteResult, error)
}
