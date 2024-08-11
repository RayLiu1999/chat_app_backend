package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Content   string             `bson:"content" json:"content"`
	Username  string             `bson:"username" json:"username"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

func (m *Message) Save(db *mongo.Database) error {
	m.CreatedAt = time.Now()
	_, err := db.Collection("messages").InsertOne(context.Background(), m)
	return err
}
