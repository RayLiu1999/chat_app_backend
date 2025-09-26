package services

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"chat_app_backend/app/repositories"
	"chat_app_backend/config"
	"chat_app_backend/utils"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type friendService struct {
	config            *config.Config
	odm               *providers.ODM
	friendRepo        repositories.FriendRepository
	userRepo          repositories.UserRepository
	fileUploadService FileUploadService // 添加 FileUploadService 依賴
	clientManager     *ClientManager
}

func NewFriendService(
	cfg *config.Config,
	odm *providers.ODM,
	friendRepo repositories.FriendRepository,
	userRepo repositories.UserRepository,
	fileUploadService FileUploadService,
	clientManager *ClientManager,
) *friendService {
	return &friendService{
		config:            cfg,
		odm:               odm,
		friendRepo:        friendRepo,
		userRepo:          userRepo,
		fileUploadService: fileUploadService,
		clientManager:     clientManager,
	}
}

// getUserPictureURL 獲取用戶頭像 URL（從 ObjectID 解析）
func (fs *friendService) getUserPictureURL(user *models.User) string {
	if user.PictureID.IsZero() || fs.fileUploadService == nil {
		return ""
	}

	pictureURL, err := fs.fileUploadService.GetFileURLByID(user.PictureID.Hex())
	if err != nil {
		return ""
	}
	return pictureURL
}

// 定義好友請求狀態的可能值
const (
	FriendStatusPending  = "pending"
	FriendStatusAccepted = "accepted"
	FriendStatusRejected = "rejected"
)

// GetFriendList 獲取好友列表
func (fs *friendService) GetFriendList(userID string) ([]models.FriendResponse, *models.MessageOptions) {
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
	var friendIds []string
	friendsStatusMap := make(map[string]string)

	for _, friend := range friends {
		var friendId string

		// 確定哪個ID是好友的ID
		if friend.UserID == userObjectID {
			friendId = friend.FriendID.Hex()
		} else {
			friendId = friend.UserID.Hex()
		}

		friendIds = append(friendIds, friendId)
		friendsStatusMap[friendId] = friend.Status
	}

	var users []models.User
	users, err = fs.userRepo.GetUserListByIds(friendIds)
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
		if fs.clientManager != nil {
			isOnline = fs.clientManager.IsUserOnline(user.ID.Hex())
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
func (fs *friendService) AddFriendRequest(userID string, username string) *models.MessageOptions {
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

	// 檢查是否已經是好友或已有請求或被封鎖
	friendQb := providers.NewQueryBuilder()
	orConditions := []bson.M{
		{"user_id": userObjectID, "friend_id": user.ID},
		{"user_id": user.ID, "friend_id": userObjectID},
	}
	friendQb.OrWhere(orConditions)

	var friend models.Friend
	err = fs.odm.FindOne(context.Background(), friendQb.GetFilter(), &friend)
	if err == nil {
		return &models.MessageOptions{Code: models.ErrFriendExists, Message: "已經有好友請求或已為好友或被封鎖"}
	}

	// 建立好友請求
	newFriend := models.Friend{
		UserID:   userObjectID,
		FriendID: user.ID,
		Status:   FriendStatusPending,
	}

	fs.odm.Create(context.Background(), &newFriend)

	return nil
}

// GetPendingRequests 獲取待處理好友請求
func (fs *friendService) GetPendingRequests(userID string) (*models.PendingRequestsResponse, *models.MessageOptions) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的用戶ID",
			Details: err.Error(),
		}
	}

	// 查詢我發送的請求
	sentQb := providers.NewQueryBuilder()
	sentQb.Where("user_id", userObjectID)
	sentQb.Where("status", FriendStatusPending)

	var sentRequests []models.Friend
	err = fs.odm.Find(context.Background(), sentQb.GetFilter(), &sentRequests)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取發送的請求失敗",
			Details: err.Error(),
		}
	}

	// 查詢我收到的請求
	receivedQb := providers.NewQueryBuilder()
	receivedQb.Where("friend_id", userObjectID)
	receivedQb.Where("status", FriendStatusPending)

	var receivedRequests []models.Friend
	err = fs.odm.Find(context.Background(), receivedQb.GetFilter(), &receivedRequests)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取收到的請求失敗",
			Details: err.Error(),
		}
	}

	// 轉換為響應格式
	var sentPending []models.PendingFriendRequest
	for _, req := range sentRequests {
		user, err := fs.userRepo.GetUserById(req.FriendID.Hex())
		if err != nil {
			continue
		}

		sentPending = append(sentPending, models.PendingFriendRequest{
			RequestID:  req.ID.Hex(),
			UserID:     user.ID.Hex(),
			Username:   user.Username,
			Nickname:   user.Nickname,
			PictureURL: fs.getUserPictureURL(user),
			SentAt:     req.CreatedAt.UnixMilli(),
			Type:       "sent",
		})
	}

	var receivedPending []models.PendingFriendRequest
	for _, req := range receivedRequests {
		user, err := fs.userRepo.GetUserById(req.UserID.Hex())
		if err != nil {
			continue
		}

		receivedPending = append(receivedPending, models.PendingFriendRequest{
			RequestID:  req.ID.Hex(),
			UserID:     user.ID.Hex(),
			Username:   user.Username,
			Nickname:   user.Nickname,
			PictureURL: fs.getUserPictureURL(user),
			SentAt:     req.CreatedAt.UnixMilli(),
			Type:       "received",
		})
	}

	response := &models.PendingRequestsResponse{
		Sent:     sentPending,
		Received: receivedPending,
	}
	response.Count.Sent = len(sentPending)
	response.Count.Received = len(receivedPending)
	response.Count.Total = response.Count.Sent + response.Count.Received

	return response, nil
}

// GetBlockedUsers 獲取封鎖用戶列表
func (fs *friendService) GetBlockedUsers(userID string) ([]models.BlockedUserResponse, *models.MessageOptions) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的用戶ID",
			Details: err.Error(),
		}
	}

	// 查詢被我封鎖的用戶
	qb := providers.NewQueryBuilder()
	qb.Where("user_id", userObjectID)
	qb.Where("status", "blocked")

	var blockedFriends []models.Friend
	err = fs.odm.Find(context.Background(), qb.GetFilter(), &blockedFriends)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取封鎖列表失敗",
			Details: err.Error(),
		}
	}

	var blockedUsers []models.BlockedUserResponse
	for _, blocked := range blockedFriends {
		user, err := fs.userRepo.GetUserById(blocked.FriendID.Hex())
		if err != nil {
			continue
		}

		blockedUsers = append(blockedUsers, models.BlockedUserResponse{
			UserID:     user.ID.Hex(),
			Username:   user.Username,
			Nickname:   user.Nickname,
			PictureURL: fs.getUserPictureURL(user),
			BlockedAt:  blocked.UpdatedAt.UnixMilli(),
		})
	}

	return blockedUsers, nil
}

// AcceptFriendRequest 接受好友請求
func (fs *friendService) AcceptFriendRequest(userID string, requestID string) *models.MessageOptions {
	requestObjectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的請求ID",
			Details: err.Error(),
		}
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的用戶ID",
			Details: err.Error(),
		}
	}

	// 查找請求
	qb := providers.NewQueryBuilder()
	qb.Where("_id", requestObjectID)
	qb.Where("friend_id", userObjectID)
	qb.Where("status", FriendStatusPending)

	utils.PrettyPrint("qb.GetFilter(): ", qb.GetFilter())

	var friendRequest models.Friend
	err = fs.odm.FindOne(context.Background(), qb.GetFilter(), &friendRequest)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "找不到待處理的好友請求",
			Details: err.Error(),
		}
	}

	// 更新狀態為接受
	friendRequest.Status = FriendStatusAccepted
	err = fs.odm.Update(context.Background(), &friendRequest)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "接受好友請求失敗",
			Details: err.Error(),
		}
	}

	return nil
}

// DeclineFriendRequest 拒絕好友請求
func (fs *friendService) DeclineFriendRequest(userID string, requestID string) *models.MessageOptions {
	requestObjectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的請求ID",
			Details: err.Error(),
		}
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的用戶ID",
			Details: err.Error(),
		}
	}

	// 查找並刪除請求
	qb := providers.NewQueryBuilder()
	qb.Where("_id", requestObjectID)
	qb.Where("friend_id", userObjectID)
	qb.Where("status", FriendStatusPending)

	var friendRequest models.Friend
	err = fs.odm.FindOne(context.Background(), qb.GetFilter(), &friendRequest)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "找不到待處理的好友請求",
			Details: err.Error(),
		}
	}

	err = fs.odm.Delete(context.Background(), &friendRequest)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "拒絕好友請求失敗",
			Details: err.Error(),
		}
	}

	return nil
}

// CancelFriendRequest 取消好友請求
func (fs *friendService) CancelFriendRequest(userID string, requestID string) *models.MessageOptions {
	requestObjectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的請求ID",
			Details: err.Error(),
		}
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的用戶ID",
			Details: err.Error(),
		}
	}

	// 查找並刪除我發送的請求
	qb := providers.NewQueryBuilder()
	qb.Where("_id", requestObjectID)
	qb.Where("user_id", userObjectID)
	qb.Where("status", FriendStatusPending)

	var friendRequest models.Friend
	err = fs.odm.FindOne(context.Background(), qb.GetFilter(), &friendRequest)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "找不到要取消的好友請求",
			Details: err.Error(),
		}
	}

	err = fs.odm.Delete(context.Background(), &friendRequest)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "取消好友請求失敗",
			Details: err.Error(),
		}
	}

	return nil
}

// BlockUser 封鎖用戶
func (fs *friendService) BlockUser(userID string, targetUserID string) *models.MessageOptions {
	// targetUserID不能為自己
	if userID == targetUserID {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "不能封鎖自己",
		}
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的用戶ID",
			Details: err.Error(),
		}
	}

	targetObjectID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的目標用戶ID",
			Details: err.Error(),
		}
	}

	// 檢查目標用戶是否存在
	_, err = fs.userRepo.GetUserById(targetUserID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "目標用戶不存在",
			Details: err.Error(),
		}
	}

	// 查找現有關係
	qb := providers.NewQueryBuilder()
	orConditions := []bson.M{
		{"user_id": userObjectID, "friend_id": targetObjectID},
		{"friend_id": userObjectID, "user_id": targetObjectID},
	}
	qb.OrWhere(orConditions)

	var existingFriend models.Friend
	err = fs.odm.FindOne(context.Background(), qb.GetFilter(), &existingFriend)

	if err != nil {
		// 如果沒有現有關係，創建新的封鎖關係
		blockedFriend := models.Friend{
			UserID:   userObjectID,
			FriendID: targetObjectID,
			Status:   "blocked",
		}

		err = fs.odm.Create(context.Background(), &blockedFriend)
		if err != nil {
			return &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "封鎖用戶失敗",
				Details: err.Error(),
			}
		}
	} else {
		// 更新現有關係為封鎖
		existingFriend.Status = "blocked"

		err = fs.odm.Update(context.Background(), &existingFriend)
		if err != nil {
			return &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "更新封鎖狀態失敗",
				Details: err.Error(),
			}
		}
	}

	return nil
}

// UnblockUser 解除封鎖用戶
func (fs *friendService) UnblockUser(userID string, targetUserID string) *models.MessageOptions {
	// targetUserID不能為自己
	if userID == targetUserID {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "不能解除封鎖自己",
		}
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的用戶ID",
			Details: err.Error(),
		}
	}

	targetObjectID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的目標用戶ID",
			Details: err.Error(),
		}
	}

	// 查找並刪除封鎖關係
	qb := providers.NewQueryBuilder()
	orConditions := []bson.M{
		{"user_id": userObjectID, "friend_id": targetObjectID, "status": "blocked"},
		{"friend_id": userObjectID, "user_id": targetObjectID, "status": "blocked"},
	}
	qb.OrWhere(orConditions)

	var blockedFriend models.Friend
	err = fs.odm.FindOne(context.Background(), qb.GetFilter(), &blockedFriend)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "找不到要解除的封鎖關係",
			Details: err.Error(),
		}
	}

	err = fs.odm.Delete(context.Background(), &blockedFriend)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "解除封鎖失敗",
			Details: err.Error(),
		}
	}

	return nil
}

// RemoveFriend 刪除好友
func (fs *friendService) RemoveFriend(userID string, friendID string) *models.MessageOptions {
	// 檢查userID和friendID是否相同
	if userID == friendID {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無法刪除自己",
		}
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的用戶ID",
			Details: err.Error(),
		}
	}

	friendObjectID, err := primitive.ObjectIDFromHex(friendID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的好友ID",
			Details: err.Error(),
		}
	}

	// 檢查目標用戶是否存在
	_, err = fs.userRepo.GetUserById(friendID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "好友不存在",
			Details: err.Error(),
		}
	}

	// 查找並刪除雙向好友關係
	// 首先查找從 userID 到 friendID 的關係
	qb1 := providers.NewQueryBuilder()
	qb1.Where("user_id", userObjectID)
	qb1.Where("friend_id", friendObjectID)
	qb1.Where("status", FriendStatusAccepted)

	var friendRelation1 models.Friend
	err1 := fs.odm.FindOne(context.Background(), qb1.GetFilter(), &friendRelation1)

	// 查找從 friendID 到 userID 的關係
	qb2 := providers.NewQueryBuilder()
	qb2.Where("user_id", friendObjectID)
	qb2.Where("friend_id", userObjectID)
	qb2.Where("status", FriendStatusAccepted)

	var friendRelation2 models.Friend
	err2 := fs.odm.FindOne(context.Background(), qb2.GetFilter(), &friendRelation2)

	// 如果兩個關係都不存在，表示不是好友
	if err1 != nil && err2 != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "好友關係不存在",
			Details: "未找到好友關係",
		}
	}

	// 刪除找到的好友關係
	if err1 == nil {
		err = fs.odm.Delete(context.Background(), &friendRelation1)
		if err != nil {
			return &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "刪除好友關係失敗",
				Details: err.Error(),
			}
		}
	}

	if err2 == nil {
		err = fs.odm.Delete(context.Background(), &friendRelation2)
		if err != nil {
			return &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "刪除好友關係失敗",
				Details: err.Error(),
			}
		}
	}

	return nil
}
