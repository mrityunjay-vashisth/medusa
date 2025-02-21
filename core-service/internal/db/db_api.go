package db

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBType string

const (
	MongoDB DBType = "mongodb"
)

type dbOptions struct {
	databaseName   string
	collectionName string
}

type DBOption func(*dbOptions)

type DBConfig struct {
	Type           DBType
	URI            string
	DatabaseName   string
	CollectionName string
}

type DBClientInterface interface {
	Connect(ctx context.Context) error
	Create(ctx context.Context, data map[string]interface{}, opts ...DBOption) (interface{}, error)
	Read(ctx context.Context, data map[string]interface{}, opts ...DBOption) (interface{}, error)
	ReadAll(ctx context.Context, data map[string]interface{}, opts ...DBOption) (interface{}, error)
	Delete(ctx context.Context, data map[string]interface{}, opts ...DBOption) (interface{}, error)
}

type mongoClient struct {
	client *mongo.Client
}

type DBClient struct {
	config      DBConfig
	mongoClient *mongoClient
}

func NewDBClient(config DBConfig) *DBClient {
	return &DBClient{
		config: config,
		mongoClient: &mongoClient{
			client: nil,
		},
	}
}

func WithDatabaseName(databaseName string) DBOption {
	return func(o *dbOptions) {
		o.databaseName = databaseName
	}
}

func WithCollectionName(collectionName string) DBOption {
	return func(o *dbOptions) {
		o.collectionName = collectionName
	}
}

func (d *DBClient) Connect(ctx context.Context) error {
	switch d.config.Type {
	case MongoDB:
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(d.config.URI))
		if err != nil {
			return err
		}
		d.mongoClient.client = client
	default:
		return errors.New("unsupported database type")
	}
	return nil
}

func (d *DBClient) Create(ctx context.Context, data map[string]interface{}, opts ...DBOption) (interface{}, error) {
	switch d.config.Type {
	case MongoDB:
		result, err := d.mongoClient.create(ctx, data, opts...)
		return result, err
	default:
		return nil, errors.New("unsupported database type")
	}
}

func (d *DBClient) Read(ctx context.Context, data map[string]interface{}, opts ...DBOption) (interface{}, error) {
	switch d.config.Type {
	case MongoDB:
		result, err := d.mongoClient.read(ctx, data, opts...)
		return result, err
	default:
		return nil, errors.New("unsupported database type")
	}
}

func (d *DBClient) ReadAll(ctx context.Context, data map[string]interface{}, opts ...DBOption) (interface{}, error) {
	switch d.config.Type {
	case MongoDB:
		result, err := d.mongoClient.readall(ctx, data, opts...)
		return result, err
	default:
		return nil, errors.New("unsupported database type")
	}
}

func (d *DBClient) Delete(ctx context.Context, data map[string]interface{}, opts ...DBOption) (interface{}, error) {
	switch d.config.Type {
	case MongoDB:
		result, err := d.mongoClient.delete(ctx, data, opts...)
		return result, err
	default:
		return nil, errors.New("unsupported database type")
	}
}
