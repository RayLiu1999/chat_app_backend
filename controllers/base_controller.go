package controllers

import (
	"chat_app_backend/database"

	"go.mongodb.org/mongo-driver/mongo"
)

var MongoConnect *mongo.Database

func init() {
	var err error
	db, err = database.ConnectDatabase()
	if err != nil {
		panic(err)
	}
}

type BaseController struct {
}
