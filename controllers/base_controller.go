package controllers

import (
	"chat_app_backend/providers"
)

var MongoConnect = providers.DBConnect()

type BaseController struct {
}
