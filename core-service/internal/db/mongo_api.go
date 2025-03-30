package db

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Create inserts a document into the specified collection.
func (m *mongoClient) create(ctx context.Context, data map[string]interface{}, opts ...DBOption) (interface{}, error) {
	dbName, collName := m.getDatabaseAndCollection(opts...)

	collection := m.client.Database(dbName).Collection(collName)
	result, err := collection.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

// FindOne retrieves a single document that matches the filter.
func (m *mongoClient) read(ctx context.Context, filter bson.M, opts ...DBOption) (map[string]interface{}, error) {
	dbName, collName := m.getDatabaseAndCollection(opts...)

	collection := m.client.Database(dbName).Collection(collName)
	var result map[string]interface{}
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil // No match found
	} else if err != nil {
		return nil, err
	}
	return result, nil
}

// Find retrieves multiple documents that match the filter.
func (m *mongoClient) readall(ctx context.Context, filter bson.M, opts ...DBOption) ([]map[string]interface{}, error) {
	dbName, collName := m.getDatabaseAndCollection(opts...)

	collection := m.client.Database(dbName).Collection(collName)
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// Delete removes one or more documents that match the filter.
func (m *mongoClient) delete(ctx context.Context, filter bson.M, opts ...DBOption) (int64, error) {
	dbName, collName := m.getDatabaseAndCollection(opts...)

	collection := m.client.Database(dbName).Collection(collName)
	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// UpdateOne updates a single document in the specified collection.
func (m *mongoClient) updateOne(ctx context.Context, filter bson.M, update bson.M, opts ...DBOption) (int64, error) {
	dbName, collName := m.getDatabaseAndCollection(opts...)

	collection := m.client.Database(dbName).Collection(collName)

	// MongoDB requires update operations to use operators like $set
	// Check if the update already has operators, if not, wrap it in $set
	if len(update) > 0 && !hasUpdateOperators(update) {
		update = bson.M{"$set": update}
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

// hasUpdateOperators checks if the update document already contains MongoDB update operators
func hasUpdateOperators(update bson.M) bool {
	for key := range update {
		if strings.HasPrefix(key, "$") {
			return true
		}
	}
	return false
}

// getDatabaseAndCollection extracts database and collection overrides if provided.
func (m *mongoClient) getDatabaseAndCollection(opts ...DBOption) (string, string) {
	userOpts := &dbOptions{}
	for _, opt := range opts {
		opt(userOpts)
	}

	dbName := userOpts.databaseName
	if dbName == "" {
		dbName = "default_db" // Fallback default
	}

	collName := userOpts.collectionName
	if collName == "" {
		collName = "default_collection" // Fallback default
	}

	return dbName, collName
}
