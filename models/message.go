package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Message struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Content   string             `json:"content" bson:"content"`
	Username  string             `json:"username" bson:"username"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

func (m *Message) Save(db *mongo.Database) error {
	m.CreatedAt = time.Now()
	_, err := MongoConnect.Collection("messages").InsertOne(context.Background(), m)
	return err
}
