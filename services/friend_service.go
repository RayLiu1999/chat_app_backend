package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"chat_app_backend/repositories"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FriendService struct {
	config            *config.Config
	friendRepo        repositories.FriendRepositoryInterface
	userRepo          repositories.UserRepositoryInterface
	userService       UserServiceInterface       // 添加 UserService 來查詢在線狀態
	fileUploadService FileUploadServiceInterface // 添加 FileUploadService 依賴
	odm               *providers.ODM
}

func NewFriendService(cfg *config.Config, odm *providers.ODM, friendRepo repositories.FriendRepositoryInterface, userRepo repositories.UserRepositoryInterface, userService UserServiceInterface, fileUploadService FileUploadServiceInterface) *FriendService {
	return &FriendService{
		config:            cfg,
		friendRepo:        friendRepo,
		userRepo:          userRepo,
		userService:       userService,
		fileUploadService: fileUploadService,
		odm:               odm,
	}
}

// getUserPictureURL 獲取用戶頭像 URL（從 ObjectID 解析）
func (fs *FriendService) getUserPictureURL(user *models.User) string {
	if user.PictureID.IsZero() || fs.fileUploadService == nil {
		return ""
	}

	pictureURL, err := fs.fileUploadService.GetFileURLByID(user.PictureID.Hex())
	if err != nil {
		return ""
	}
	return pictureURL
}

func (fs *FriendService) GetFriendById(userID string) (*models.Friend, *models.MessageOptions) {
	friend, err := fs.friendRepo.GetFriendById(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "獲取好友信息失敗",
		}
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
func (fs *FriendService) GetFriendList(userID string) ([]models.FriendResponse, *models.MessageOptions) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Details: err,
			Message: "無效的用戶ID格式",
		}
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
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "查詢好友列表失敗",
		}
	}

	if len(friends) == 0 {
		return []models.FriendResponse{}, nil
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
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "獲取好友用戶信息失敗",
		}
	}

	var apiFriend []models.FriendResponse
	for _, user := range users {
		status := friendsStatusMap[user.ID.Hex()]
		// 查詢好友的在線狀態
		isOnline := false
		if fs.userService != nil {
			isOnline = fs.userService.IsUserOnlineByWebSocket(user.ID.Hex())
		}

		apiFriend = append(apiFriend, models.FriendResponse{
			ID:         user.ID.Hex(),
			Name:       user.Username,
			Nickname:   user.Nickname,
			PictureURL: fs.getUserPictureURL(&user),
			Status:     status,
			IsOnline:   isOnline, // 添加在線狀態
		})
	}

	return apiFriend, nil
}

// AddFriendRequest 發送好友請求
func (fs *FriendService) AddFriendRequest(userID string, username string) *models.MessageOptions {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &models.MessageOptions{Code: models.ErrInvalidParams, Message: "無效的用戶ID", Details: err.Error()}
	}

	// 檢查好友是否存在
	qb := providers.NewQueryBuilder()
	qb.Where("username", username)

	var user models.User
	err = fs.odm.FindOne(context.Background(), qb.GetFilter(), &user)
	if err != nil {
		return &models.MessageOptions{Code: models.ErrUserNotFound, Message: "好友不存在", Details: err.Error()}
	}

	// 不能加自己為好友
	if userObjectID == user.ID {
		return &models.MessageOptions{Code: models.ErrInvalidParams, Message: "不能加自己為好友"}
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
		return &models.MessageOptions{Code: models.ErrFriendExists, Message: "已經有好友請求或已為好友"}
	}

	// 建立好友請求
	newFriend := models.Friend{
		UserID:   userObjectID,
		FriendID: user.ID,
		Status:   FriendStatusPending,
	}

	fs.odm.Create(context.Background(), &newFriend)

	return &models.MessageOptions{Message: "好友請求已發送"}
}

// UpdateFriendStatus 更新好友狀態
func (fs *FriendService) UpdateFriendStatus(userID string, friendID string, status string) *models.MessageOptions {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Details: err,
			Message: "無效的用戶ID格式",
		}
	}

	// 檢查好友是否存在
	var user models.User
	err = fs.odm.FindByID(context.Background(), friendID, &user)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrUserNotFound,
			Details: err,
			Message: "好友不存在",
		}
	}

	// 不能加自己為好友
	if userObjectID == user.ID {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "不能對自己執行此操作",
		}
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
		return &models.MessageOptions{
			Code:    models.ErrFriendExists,
			Message: "已經是好友",
		}
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
			return &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Details: err,
				Message: "找不到待處理的好友請求",
			}
		}

		err = fs.odm.Delete(context.Background(), &friendToDelete)
		if err != nil {
			return &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Details: err,
				Message: "刪除好友請求失敗",
			}
		}
		return nil
	}

	// 更新好友狀態
	updateQb := providers.NewQueryBuilder()
	updateQb.Where("user_id", user.ID).
		Where("friend_id", userObjectID).
		Where("status", FriendStatusPending)

	var friendToUpdate models.Friend
	err = fs.odm.FindOne(context.Background(), updateQb.GetFilter(), &friendToUpdate)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "找不到待處理的好友請求",
		}
	}

	friendToUpdate.Status = status
	err = fs.odm.Update(context.Background(), &friendToUpdate)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "更新好友狀態失敗",
		}
	}

	return nil
}
