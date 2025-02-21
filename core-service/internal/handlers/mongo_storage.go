package handlers

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// **MongoStorage implements the Storage interface**
type MongoStorage struct {
	client *mongo.Client
	dbName string
}

// **NewMongoStorage creates a new MongoDB-backed storage**
func NewMongoStorage(client *mongo.Client, dbName string) *MongoStorage {
	return &MongoStorage{client: client, dbName: dbName}
}

// **Find Implementation**
func (m *MongoStorage) Find(ctx context.Context, collection string, filter bson.M) (*mongo.Cursor, error) {
	return m.client.Database(m.dbName).Collection(collection).Find(ctx, filter)
}

// **FindOne Implementation**
func (m *MongoStorage) FindOne(ctx context.Context, collection string, filter bson.M) (*mongo.SingleResult, error) {
	return m.client.Database(m.dbName).Collection(collection).FindOne(ctx, filter), nil
}

// **InsertOne Implementation**
func (m *MongoStorage) InsertOne(ctx context.Context, collection string, document interface{}) (*mongo.InsertOneResult, error) {
	return m.client.Database(m.dbName).Collection(collection).InsertOne(ctx, document)
}

// **DeleteOne Implementation**
func (m *MongoStorage) DeleteOne(ctx context.Context, collection string, filter bson.M) (*mongo.DeleteResult, error) {
	return m.client.Database(m.dbName).Collection(collection).DeleteOne(ctx, filter)
}
