package db

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestCreateDocument(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Create document", func(mt *mtest.T) {
		// **Simulate MongoDB response**
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// **Initialize DBClient with mock MongoDB**
		client := NewDBClient(DBConfig{
			Type:           MongoDB,
			URI:            "mongodb://fake-uri", // Not actually used in mocks
			DatabaseName:   "test_db",
			CollectionName: "test_collection",
		})
		client.mongoClient.client = mt.Client

		// **Call actual Create API**
		ctx := context.Background()
		data := map[string]interface{}{"name": "TestUser", "email": "test@example.com"}
		insertedID, err := client.Create(ctx, data)

		// **Assertions**
		assert.NoError(t, err, "Create should not return an error")
		assert.NotNil(t, insertedID, "Inserted ID should not be nil")
	})
}

func TestReadDocument(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Read document", func(mt *mtest.T) {
		// **Simulate MongoDB response**
		expectedDoc := bson.D{{Key: "name", Value: "TestUser"}}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test_db.test_collection", mtest.FirstBatch, expectedDoc))

		client := NewDBClient(DBConfig{
			Type:           MongoDB,
			URI:            "mongodb://fake-uri",
			DatabaseName:   "test_db",
			CollectionName: "test_collection",
		})
		client.mongoClient.client = mt.Client

		// **Call actual Read API**
		ctx := context.Background()
		filter := bson.M{"name": "TestUser"}
		result, err := client.Read(ctx, filter)

		// **Assertions**
		assert.NoError(t, err, "Read should not return an error")
		assert.NotNil(t, result, "Result should not be nil")
		jsonData, err := json.Marshal(result)
		assert.NoError(t, err, "JSON conversion should not return an error")

		// **Expected JSON output**
		expectedJSON := `{"name":"TestUser"}`
		assert.JSONEq(t, expectedJSON, string(jsonData), "JSON output should match")
		// assert.Equal(t, "TestUser", result["name"]., "Name should match")
	})
}

func TestDeleteDocument(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Delete document", func(mt *mtest.T) {
		// **Simulate a successful delete response from MongoDB**
		deleteResponse := mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)})
		mt.AddMockResponses(deleteResponse)

		client := NewDBClient(DBConfig{
			Type:           MongoDB,
			URI:            "mongodb://fake-uri",
			DatabaseName:   "test_db",
			CollectionName: "test_collection",
		})
		client.mongoClient.client = mt.Client

		// **Call actual Delete API**
		ctx := context.Background()
		filter := bson.M{"name": "TestUser"}
		deletedCount, err := client.Delete(ctx, filter)

		// **Assertions**
		assert.NoError(t, err, "Delete should not return an error")
		assert.GreaterOrEqual(t, deletedCount, int64(1), "Should delete at least one document")
	})
}
