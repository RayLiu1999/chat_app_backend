package models

import (
	"chat_app_backend/database"
)

var MongoConnect = database.MongoDBConnect()

type Base struct {
}
