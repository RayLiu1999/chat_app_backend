package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
	"chat_app_backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FriendController struct {
	config        *config.Config
	mongoConnect  *mongo.Database
	friendService services.FriendServiceInterface
}

func NewFriendController(cfg *config.Config, mongodb *mongo.Database, friendService services.FriendServiceInterface) *FriendController {
	return &FriendController{
		config:        cfg,
		mongoConnect:  mongodb,
		friendService: friendService,
	}
}

type APIFriend struct {
	ID       string `json:"id" bson:"_id"`
	Name     string `json:"name" bson:"name"`
	Nickname string `json:"nickname" bson:"nickname"`
	Picture  string `json:"picture" bson:"picture"`
	Status   string `json:"status" bson:"status"`
}

// 更新好友請求結構
type FriendRequestStatus struct {
	Status string `json:"status" binding:"required,oneof=pending accepted rejected"`
}

// 定義好友請求狀態的可能值
const (
	FriendStatusPending  = "pending"
	FriendStatusAccepted = "accepted"
	FriendStatusRejected = "rejected"
)

// 取得用戶資訊
func (uc *FriendController) GetFriendList(c *gin.Context) {
	_, objectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized, Displayable: true})
		return
	}

	// 從 friends 集合獲取朋友和邀請資料
	collection := uc.mongoConnect.Collection("friends")

	// 查詢與當前用戶相關的所有好友關係（包括發送和接收的請求）
	// filter := bson.M{"friend_id": objectID}
	filter := bson.M{
		"$or": []bson.M{
			{"user_id": objectID, "status": FriendStatusAccepted},
			{"friend_id": objectID, "status": FriendStatusAccepted},
		},
	}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Displayable: true})
		return
	}
	defer cursor.Close(context.Background())

	// 存儲好友關係和對應的狀態
	var friends []models.Friend
	var friendIds []primitive.ObjectID
	friendsStatusMap := make(map[string]string) // 用於存儲每個好友ID對應的狀態

	// 解析好友關係數據
	err = cursor.All(context.Background(), &friends)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Displayable: true})
		return
	}

	// 處理好友ID和狀態
	for _, friend := range friends {
		var friendId primitive.ObjectID

		// 確定哪個ID是好友的ID
		if friend.UserID == objectID {
			friendId = friend.FriendID
		} else {
			friendId = friend.UserID
		}

		// 將好友ID添加到列表中
		friendIds = append(friendIds, friendId)

		// 存儲好友狀態
		friendsStatusMap[friendId.Hex()] = friend.Status
	}

	// 如果沒有好友關係，直接返回空列表
	if len(friendIds) == 0 {
		utils.SuccessResponse(c, []APIFriend{}, utils.MessageOptions{Message: "好友資訊獲取成功"})
		return
	}

	// 用取得的friend_id陣列取得好友資訊
	var userCollection = uc.mongoConnect.Collection("users")
	var apiFriend []APIFriend

	// 用where in friend_ids取得使用者資訊
	friendRows, err := userCollection.Find(context.Background(), bson.M{"_id": bson.M{"$in": friendIds}})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: "伺服器內部錯誤"})
		return
	}
	defer friendRows.Close(context.Background())

	// 解析用戶數據並組合API響應
	for friendRows.Next(context.Background()) {
		var user models.User
		err := friendRows.Decode(&user)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: "伺服器內部錯誤"})
			return
		}

		// 從map中獲取該好友的狀態
		status := friendsStatusMap[user.ID.Hex()]

		apiFriend = append(apiFriend, APIFriend{
			ID:       user.ID.Hex(),
			Name:     user.Username,
			Nickname: user.Nickname,
			Picture:  user.Picture,
			Status:   status,
		})
	}

	utils.SuccessResponse(c, apiFriend, utils.MessageOptions{Message: "好友資訊獲取成功"})
}

// 建立好友請求
func (uc *FriendController) AddFriendRequest(c *gin.Context) {
	_, objectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized, Message: "未授權的請求"})
		return
	}

	username := c.Param("username")

	// 檢查好友是否存在
	collection := uc.mongoConnect.Collection("users")
	var user models.User
	err = collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.MessageOptions{Code: utils.ErrUserNotFound, Message: "好友不存在"})
		return
	}

	// 不能加自己為好友
	if objectID == user.ID {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams, Message: "不能加自己為好友"})
		return
	}

	// 檢查是否已經是好友
	collection = uc.mongoConnect.Collection("friends")
	var friend models.Friend

	// 檢查雙方是否已經是好友（在任一方向）
	filter := bson.M{
		"$or": []bson.M{
			{"user_id": objectID, "friend_id": user.ID, "status": FriendStatusAccepted},
			{"user_id": user.ID, "friend_id": objectID, "status": FriendStatusAccepted},
		},
	}
	err = collection.FindOne(context.Background(), filter).Decode(&friend)
	if err == nil {
		utils.ErrorResponse(c, http.StatusConflict, utils.MessageOptions{Code: utils.ErrFriendExists, Message: "已經有好友請求或已為好友"})
		return
	}

	// 建立好友請求
	friend = models.Friend{
		UserID:    objectID,
		FriendID:  user.ID,
		Status:    FriendStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = collection.InsertOne(context.Background(), friend)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: "伺服器內部錯誤"})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "好友請求已發送"})
}

// 更新好友狀態
func (uc *FriendController) UpdateFriendStatus(c *gin.Context) {
	_, objectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	var friend_id primitive.ObjectID
	friend_id, err = primitive.ObjectIDFromHex(c.Param("friend_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams, Message: "好友ID格式錯誤"})
		return
	}

	// 取得put中status資料
	var requestBody FriendRequestStatus
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}
	status := requestBody.Status
	fmt.Println(status)

	// 檢查好友是否存在
	collection := uc.mongoConnect.Collection("users")
	var user models.User
	err = collection.FindOne(context.Background(), bson.M{"_id": friend_id}).Decode(&user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.MessageOptions{Code: utils.ErrUserNotFound, Message: "好友不存在"})
		return
	}

	// 不能加自己為好友
	if objectID == user.ID {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams, Message: "不能加自己為好友"})
		return
	}

	// 檢查是否已經是好友
	collection = uc.mongoConnect.Collection("friends")
	var friend models.Friend

	// 檢查雙方是否已經是好友（在任一方向）
	filter := bson.M{
		"$or": []bson.M{
			{"user_id": objectID, "friend_id": user.ID, "status": FriendStatusAccepted},
			{"user_id": user.ID, "friend_id": objectID, "status": FriendStatusAccepted},
		},
	}
	err = collection.FindOne(context.Background(), filter).Decode(&friend)
	if err == nil {
		utils.ErrorResponse(c, http.StatusConflict, utils.MessageOptions{Code: utils.ErrFriendExists, Message: "已經是好友"})
		return
	}

	// 拒絕則刪除請求紀錄
	if status == FriendStatusRejected {
		_, err = collection.DeleteOne(context.Background(), bson.D{{"user_id", user.ID}, {"friend_id", objectID}, {"status", FriendStatusPending}})
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
			return
		}
		utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "已拒絕好友請求"})
		return
	}

	// 更新好友狀態
	friend = models.Friend{
		UserID:    objectID,
		FriendID:  user.ID,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = collection.UpdateOne(context.Background(), bson.D{{"user_id", user.ID}, {"friend_id", objectID}, {"status", FriendStatusPending}}, bson.M{"$set": friend})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "已接受好友請求"})
}
