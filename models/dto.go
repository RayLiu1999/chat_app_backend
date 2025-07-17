package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserResponse struct {
	ID       string `json:"id" bson:"_id"`
	Username string `json:"username" bson:"username"`
	Email    string `json:"email" bson:"email"`
	Nickname string `json:"nickname" bson:"nickname"`
	Picture  string `json:"picture" bson:"picture"`
}

// LoginResponse 包含登入成功後返回的資訊
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
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

type DMRoomResponse struct {
	RoomID    primitive.ObjectID `json:"room_id" bson:"room_id"`
	Nickname  string             `json:"nickname" bson:"nickname"`
	Picture   string             `json:"picture" bson:"picture"`
	Timestamp int64              `json:"timestamp" bson:"timestamp"`
}

type MessageResponse struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	RoomType  RoomType           `json:"room_type" bson:"room_type"`
	RoomID    string             `json:"room_id" bson:"room_id"`
	SenderID  string             `json:"sender_id" bson:"sender_id"`
	Content   string             `json:"content" bson:"content"`
	Timestamp int64              `json:"timestamp" bson:"timestamp"`
}

type APIFriend struct {
	ID       string `json:"id" bson:"_id"`
	Name     string `json:"name" bson:"name"`
	Nickname string `json:"nickname" bson:"nickname"`
	Picture  string `json:"picture" bson:"picture"`
	Status   string `json:"status" bson:"status"`
}

func (d *DMRoom) GetID() primitive.ObjectID {
	return d.ID
}

func (d *DMRoom) SetID(id primitive.ObjectID) {
	d.ID = id
}
