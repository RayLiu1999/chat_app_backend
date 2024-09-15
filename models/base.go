package models

import "go.mongodb.org/mongo-driver/mongo"

type Base struct {
	MongoConnect *mongo.Database
}
