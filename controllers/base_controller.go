package controllers

import (
	"chat_app_backend/database"
)

var MongoConnect = database.MongoDBConnect()

type BaseController struct {
}
