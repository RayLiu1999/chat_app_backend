package models

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username string             `bson:"username" json:"username"`
	Password string             `bson:"password" json:"password,omitempty"`
}

func (u *User) Register(db *mongo.Database) error {
	var existingUser User
	err := db.Collection("users").FindOne(context.Background(), bson.M{"username": u.Username}).Decode(&existingUser)
	if err == nil {
		return errors.New("username already exists")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	u.Password = string(hashedPassword)

	result, err := db.Collection("users").InsertOne(context.Background(), u)
	if err != nil {
		return err
	}

	u.ID = result.InsertedID.(primitive.ObjectID)
	u.Password = "" // Don't return the password
	return nil
}

func AuthenticateUser(db *mongo.Database, username, password string) (*User, error) {
	var user User
	err := db.Collection("users").FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	user.Password = "" // Don't return the password
	return &user, nil
}
