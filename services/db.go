package services

import (
	"context"
	"time"

	"chat_app_backend/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	client *mongo.Client
}

func (db *Database) Client() *mongo.Client {
	return db.client
}

func (db *Database) Context() context.Context {
	return context.TODO() // 或者返回其他 context，例如 context.Background()
}

func InitDB() (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoURI))
	if err != nil {
		return nil, err
	}

	return client.Database(config.DBName), nil
}
