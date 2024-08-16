package controllers

import "go.mongodb.org/mongo-driver/mongo"

type BaseController struct {
	DB *mongo.Database
}
