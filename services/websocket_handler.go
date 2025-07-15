package services

import (
	"chat_app_backend/models"
	"chat_app_backend/utils"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// WebSocketHandler 處理 WebSocket 相關操作
type WebSocketHandler struct {
	mongoConnect   *mongo.Database
	clientManager  *ClientManager
	roomManager    *RoomManager
	messageHandler *MessageHandler
}

// NewWebSocketHandler 創建新的 WebSocket 處理器
func NewWebSocketHandler(mongoConnect *mongo.Database, clientManager *ClientManager, roomManager *RoomManager, messageHandler *MessageHandler) *WebSocketHandler {
	return &WebSocketHandler{
		mongoConnect:   mongoConnect,
		clientManager:  clientManager,
		roomManager:    roomManager,
		messageHandler: messageHandler,
	}
}

// HandleWebSocket 處理 WebSocket 連線
func (wsh *WebSocketHandler) HandleWebSocket(ws *websocket.Conn, userID string) {
	// 初始化用戶
	client := &Client{
		UserID:        userID,
		Conn:          ws,
		Subscribed:    make(map[string]bool),
		RoomActivity:  make(map[string]time.Time),
		ActivityMutex: sync.RWMutex{},
	}
	wsh.clientManager.Register(client)

	for {
		var msg WsMessage[json.RawMessage]
		err := ws.ReadJSON(&msg)
		if err != nil {
			utils.PrettyPrintf("Read message failed: %v", err)
			wsh.clientManager.Unregister(client)
			break
		}

		utils.PrettyPrint("Received message:", msg)

		switch msg.Action {
		case "join_room":
			utils.PrettyPrintf("User %s is joining room: %s", userID, msg.Data)
			wsh.handleJoinRoom(ws, client, msg.Data)
		case "leave_room":
			wsh.handleLeaveRoom(ws, client, msg.Data)
		case "send_message":
			utils.PrettyPrintf("User %s is sending message in room: %s", userID, msg.Data)
			wsh.handleSendMessage(ws, client, msg.Data, userID)
		}
	}
}

// handleJoinRoom 處理加入房間請求
func (wsh *WebSocketHandler) handleJoinRoom(ws *websocket.Conn, client *Client, data json.RawMessage) {
	var requestData struct {
		RoomID   string          `json:"room_id"`
		RoomType models.RoomType `json:"room_type"`
	}
	err := json.Unmarshal(data, &requestData)
	if err != nil {
		utils.PrettyPrintf("Failed to parse join_room data: %v", err)
		return
	}

	type joinRoomResponse struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	allowed, err := wsh.roomManager.checkUserAllowedJoinRoom(client.UserID, requestData.RoomID, requestData.RoomType)
	if err != nil {
		message := &WsMessage[joinRoomResponse]{
			Action: "join_room",
			Data: joinRoomResponse{
				Status:  "error",
				Message: "Failed to check user allowed join room",
			},
		}
		ws.WriteJSON(message)
		return
	}
	if !allowed {
		message := &WsMessage[joinRoomResponse]{
			Action: "join_room",
			Data: joinRoomResponse{
				Status:  "error",
				Message: "User not allowed to join room",
			},
		}
		ws.WriteJSON(message)
		return
	}

	// 初始化房間
	wsh.roomManager.InitRoom(requestData.RoomType, requestData.RoomID)

	// 加入房間
	wsh.roomManager.JoinRoom(client, requestData.RoomType, requestData.RoomID)

	message := &WsMessage[joinRoomResponse]{
		Action: "join_room",
		Data: joinRoomResponse{
			Status:  "success",
			Message: "Joined " + string(requestData.RoomType) + " room " + requestData.RoomID,
		},
	}

	ws.WriteJSON(message)
}

// handleLeaveRoom 處理離開房間請求
func (wsh *WebSocketHandler) handleLeaveRoom(ws *websocket.Conn, client *Client, data json.RawMessage) {
	var requestData struct {
		RoomID   string          `json:"room_id"`
		RoomType models.RoomType `json:"room_type"`
	}
	err := json.Unmarshal(data, &requestData)
	if err != nil {
		utils.PrettyPrintf("Failed to parse leave_room data: %v", err)
		return
	}

	type leaveRoomResponse struct {
		Action  string `json:"action"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	wsh.roomManager.LeaveRoom(client, requestData.RoomType, requestData.RoomID)
	message := &WsMessage[leaveRoomResponse]{
		Action: "leave_room",
		Data: leaveRoomResponse{
			Status:  "success",
			Message: "Left " + string(requestData.RoomType) + " room " + requestData.RoomID,
		},
	}
	ws.WriteJSON(message)
}

// handleSendMessage 處理發送消息請求
func (wsh *WebSocketHandler) handleSendMessage(ws *websocket.Conn, client *Client, data json.RawMessage, userID string) {
	var requestData struct {
		RoomID   string          `json:"room_id"`
		RoomType models.RoomType `json:"room_type"`
		Content  string          `json:"content"`
	}
	err := json.Unmarshal(data, &requestData)
	if err != nil {
		utils.PrettyPrintf("Failed to parse send_message data: %v", err)
		return
	}

	// 處理私聊房間邏輯
	if requestData.RoomType == models.RoomTypeDM {
		wsh.handleDMRoomCreation(requestData.RoomID, userID)
	}

	message := &WsMessage[MessageResponse]{
		Action: "send_message",
		Data: MessageResponse{
			RoomID:    requestData.RoomID,
			RoomType:  requestData.RoomType,
			SenderID:  userID,
			Content:   requestData.Content,
			Timestamp: time.Now().UnixMilli(),
		},
	}
	wsh.messageHandler.HandleMessage(message)
}

// handleDMRoomCreation 處理私聊房間創建邏輯
func (wsh *WebSocketHandler) handleDMRoomCreation(roomID, userID string) {
	key := RoomKey{Type: models.RoomTypeDM, RoomID: roomID}
	room, exists := wsh.roomManager.GetRoom(models.RoomTypeDM, roomID)
	if !exists {
		utils.PrettyPrintf("Room %s not found", key.String())
		return
	}

	// 檢查房間是否只有一個客戶端
	room.Mutex.RLock()
	isOnlyOneClient := len(room.Clients) == 1
	room.Mutex.RUnlock()

	// 如果是私聊且只有一個客戶端，則判斷是否建立房間
	if isOnlyOneClient {
		roomObjectID, err := primitive.ObjectIDFromHex(roomID)
		if err != nil {
			utils.PrettyPrintf("Failed to parse room_id: %v", err)
			return
		}

		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			utils.PrettyPrintf("Failed to parse user_id: %v", err)
			return
		}

		// 先用RoomID找到ChatWithUserID
		var dmRoomList []models.DMRoom
		dbRoomCollection := wsh.mongoConnect.Collection("dm_rooms")
		cursor, err := dbRoomCollection.Find(context.Background(), bson.M{
			"room_id": roomObjectID,
			"$or": []bson.M{
				{"user_id": userObjectID},
				{"chat_with_user_id": userObjectID},
			},
		})
		if err != nil {
			utils.PrettyPrintf("Failed to find dm room: %v", err)
			return
		}

		cursor.All(context.Background(), &dmRoomList)

		// 如果對方房間不存在，則建立
		if len(dmRoomList) == 1 {
			for _, dmRoom := range dmRoomList {
				RoomID := dmRoom.RoomID
				dbRoomCollection.InsertOne(context.Background(), models.DMRoom{
					RoomID:         RoomID,
					UserID:         dmRoom.ChatWithUserID,
					ChatWithUserID: dmRoom.UserID,
					IsHidden:       false,
				})
			}
		}
	}
}
