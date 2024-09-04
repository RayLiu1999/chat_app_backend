package providers

import (
	"context"
	"log"
	"sync"
	"time"

	"chat_app_backend/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoConnect *mongo.Database
	once         sync.Once
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
	var err error
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoURI))
		if err != nil {
			return
		}

		mongoConnect = client.Database(config.DBName)

		log.Println("Connected to MongoDB!")
	})

	return mongoConnect, err
}

func DBConnect() *mongo.Database {
	db, err := InitDB()
	if err != nil {
		panic(err)
	}

	return db
}
