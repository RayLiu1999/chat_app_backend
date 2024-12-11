package services

import (
	"chat_app_backend/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (bs *BaseService) GetUserById(objectID primitive.ObjectID) (*models.User, error) {
	user, err := bs.repo.GetUserById(objectID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
