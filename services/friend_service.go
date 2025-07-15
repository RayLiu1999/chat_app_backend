package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"chat_app_backend/repositories"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FriendService struct {
	config       *config.Config
	mongoConnect *mongo.Database
	friendRepo   repositories.FriendRepositoryInterface
	userRepo     repositories.UserRepositoryInterface
	odm          *providers.ODM
}

func NewFriendService(cfg *config.Config, mongodb *mongo.Database, friendRepo repositories.FriendRepositoryInterface, userRepo repositories.UserRepositoryInterface) *FriendService {
	return &FriendService{
		config:       cfg,
		mongoConnect: mongodb,
		friendRepo:   friendRepo,
		userRepo:     userRepo,
		odm:          providers.NewODM(mongodb),
	}
}

func (fs *FriendService) GetFriendById(userID string) (*models.Friend, error) {
	friend, err := fs.friendRepo.GetFriendById(userID)
	if err != nil {
		return nil, err
	}

	return friend, nil
}

// 定義好友請求狀態的可能值
const (
	FriendStatusPending  = "pending"
	FriendStatusAccepted = "accepted"
	FriendStatusRejected = "rejected"
)

// GetFriendList 獲取好友列表
func (fs *FriendService) GetFriendList(userID string) ([]models.APIFriend, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 使用QueryBuilder構建查詢
	qb := providers.NewQueryBuilder()
	orConditions := []bson.M{
		{"user_id": userObjectID, "status": FriendStatusAccepted},
		{"friend_id": userObjectID, "status": FriendStatusAccepted},
	}
	qb.OrWhere(orConditions)

	var friends []models.Friend
	err = fs.odm.Find(context.Background(), qb.GetFilter(), &friends)
	if err != nil {
		return nil, err
	}

	if len(friends) == 0 {
		return []models.APIFriend{}, nil
	}

	// 處理好友ID和狀態
	var friendIds []primitive.ObjectID
	friendsStatusMap := make(map[string]string)

	for _, friend := range friends {
		var friendId primitive.ObjectID

		// 確定哪個ID是好友的ID
		if friend.UserID == userObjectID {
			friendId = friend.FriendID
		} else {
			friendId = friend.UserID
		}

		friendIds = append(friendIds, friendId)
		friendsStatusMap[friendId.Hex()] = friend.Status
	}

	// 使用QueryBuilder查詢用戶資訊
	userQb := providers.NewQueryBuilder()
	userQb.WhereIn("_id", friendIds)

	var users []models.User
	err = fs.odm.Find(context.Background(), userQb.GetFilter(), &users)
	if err != nil {
		return nil, err
	}

	var apiFriend []models.APIFriend
	for _, user := range users {
		status := friendsStatusMap[user.ID.Hex()]
		apiFriend = append(apiFriend, models.APIFriend{
			ID:       user.ID.Hex(),
			Name:     user.Username,
			Nickname: user.Nickname,
			Picture:  user.Picture,
			Status:   status,
		})
	}

	return apiFriend, nil
}

// AddFriendRequest 發送好友請求
func (fs *FriendService) AddFriendRequest(userID string, username string) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// 檢查好友是否存在
	qb := providers.NewQueryBuilder()
	qb.Where("username", username)

	var user models.User
	err = fs.odm.FindOne(context.Background(), qb.GetFilter(), &user)
	if err != nil {
		return errors.New("好友不存在")
	}

	// 不能加自己為好友
	if userObjectID == user.ID {
		return errors.New("不能加自己為好友")
	}

	// 檢查是否已經是好友
	friendQb := providers.NewQueryBuilder()
	orConditions := []bson.M{
		{"user_id": userObjectID, "friend_id": user.ID, "status": FriendStatusAccepted},
		{"user_id": user.ID, "friend_id": userObjectID, "status": FriendStatusAccepted},
	}
	friendQb.OrWhere(orConditions)

	var friend models.Friend
	err = fs.odm.FindOne(context.Background(), friendQb.GetFilter(), &friend)
	if err == nil {
		return errors.New("已經有好友請求或已為好友")
	}

	// 建立好友請求
	newFriend := models.Friend{
		UserID:   userObjectID,
		FriendID: user.ID,
		Status:   FriendStatusPending,
	}

	return fs.odm.Create(context.Background(), &newFriend)
}

// UpdateFriendStatus 更新好友狀態
func (fs *FriendService) UpdateFriendStatus(userID string, friendID string, status string) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// 檢查好友是否存在
	var user models.User
	err = fs.odm.FindByID(context.Background(), friendID, &user)
	if err != nil {
		return errors.New("好友不存在")
	}

	// 不能加自己為好友
	if userObjectID == user.ID {
		return errors.New("不能加自己為好友")
	}

	// 檢查是否已經是好友
	qb := providers.NewQueryBuilder()
	orConditions := []bson.M{
		{"user_id": userObjectID, "friend_id": user.ID, "status": FriendStatusAccepted},
		{"user_id": user.ID, "friend_id": userObjectID, "status": FriendStatusAccepted},
	}
	qb.OrWhere(orConditions)

	var friend models.Friend
	err = fs.odm.FindOne(context.Background(), qb.GetFilter(), &friend)
	if err == nil {
		return errors.New("已經是好友")
	}

	// 拒絕則刪除請求紀錄
	if status == FriendStatusRejected {
		deleteQb := providers.NewQueryBuilder()
		deleteQb.Where("user_id", user.ID).
			Where("friend_id", userObjectID).
			Where("status", FriendStatusPending)

		var friendToDelete models.Friend
		err = fs.odm.FindOne(context.Background(), deleteQb.GetFilter(), &friendToDelete)
		if err != nil {
			return err
		}

		return fs.odm.Delete(context.Background(), &friendToDelete)
	}

	// 更新好友狀態
	updateQb := providers.NewQueryBuilder()
	updateQb.Where("user_id", user.ID).
		Where("friend_id", userObjectID).
		Where("status", FriendStatusPending)

	var friendToUpdate models.Friend
	err = fs.odm.FindOne(context.Background(), updateQb.GetFilter(), &friendToUpdate)
	if err != nil {
		return err
	}

	friendToUpdate.Status = status
	return fs.odm.Update(context.Background(), &friendToUpdate)
}
