package models

import (
	"chat_app_backend/providers"
)

var MongoConnect = providers.DBConnect()

type Base struct {
}
