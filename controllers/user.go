package controllers

import (
	"context"
	"errors"

	"chat_app_backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func Register(db *mongo.Database) error {
	var user models.User
	err := db.Collection("users").FindOne(context.Background(), bson.M{"username": user.Username}).Decode(&user)
	if err == nil {
		return errors.New("username already exists")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	result, err := db.Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	user.Password = "" // Don't return the password
	return nil
}

func AuthenticateUser(db *mongo.Database, username, password string) (*models.User, error) {
	var user models.User
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

// func Register(c *gin.Context) {
// 	var user models.User
// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	models.Users = append(models.Users, user)
// 	c.JSON(http.StatusOK, gin.H{"message": "registration successful"})
// }

// func Login(c *gin.Context) {
// var loginDetails models.User
// if err := c.ShouldBindJSON(&loginDetails); err != nil {
// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 	return
// }

// for _, user := range models.Users {
// 	if user.Username == loginDetails.Username && user.Password == loginDetails.Password {
// 		token, _ := utils.GenerateToken(user.Username)
// 		c.JSON(http.StatusOK, gin.H{"token": token})
// 		return
// 	}
// }

// c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
// }
