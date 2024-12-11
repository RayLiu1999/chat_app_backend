package repositories

import (
	"chat_app_backend/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (br *BaseRepository) GetUserById(objectId primitive.ObjectID) (*models.User, error) {
	var user models.User
	var collection = br.MongoConnect.Collection("users")

	err := collection.FindOne(context.Background(), bson.M{"_id": objectId}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
