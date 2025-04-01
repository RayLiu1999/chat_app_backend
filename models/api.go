package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserResponse struct {
	Name    string `json:"name" bson:"name"`
	Picture string `json:"picture" bson:"picture"`
	Email   string `json:"email" bson:"email"`
}

type ServerResponse struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	PictureURL  string             `json:"picture_url" bson:"picture_url"`
	Description string             `json:"description" bson:"description"`
}

type ChannelResponse struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	ServerID    primitive.ObjectID `json:"server_id" bson:"server_id"`
	Name        string             `json:"name" bson:"name"`
	Type        string             `json:"type" bson:"type"`
	PictureURL  string             `json:"picture_url" bson:"picture_url"`
	Description string             `json:"description" bson:"description"`
}
