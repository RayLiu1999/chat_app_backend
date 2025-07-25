package services

import (
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"chat_app_backend/utils"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebSocketHandler 處理 WebSocket 相關操作
type WebSocketHandler struct {
	odm            *providers.ODM
	clientManager  *ClientManager
	roomManager    *RoomManager
	messageHandler *MessageHandler
	userService    UserServiceInterface
}

// NewWebSocketHandler 創建新的 WebSocket 處理器
func NewWebSocketHandler(odm *providers.ODM, clientManager *ClientManager, roomManager *RoomManager, messageHandler *MessageHandler, userService UserServiceInterface) *WebSocketHandler {
	return &WebSocketHandler{
		odm:            odm,
		clientManager:  clientManager,
		roomManager:    roomManager,
		messageHandler: messageHandler,
		userService:    userService,
	}
}

// HandleWebSocket 處理 WebSocket 連線
func (wsh *WebSocketHandler) HandleWebSocket(ws *websocket.Conn, userID string) {
	// 設置連接參數
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 初始化用戶
	client := &Client{
		UserID:        userID,
		Conn:          ws,
		Subscribed:    make(map[string]bool),
		RoomActivity:  make(map[string]time.Time),
		ActivityMutex: sync.RWMutex{},
	}
	wsh.clientManager.Register(client)

	// 設置用戶為在線狀態
	if err := wsh.userService.SetUserOnline(userID); err != nil {
		utils.PrettyPrintf("Failed to set user %s online: %v", userID, err)
	}

	// 啟動心跳檢測
	go wsh.pingHandler(ws, client)

	// 處理消息循環
	defer func() {
		if r := recover(); r != nil {
			utils.PrettyPrintf("WebSocket handler panic recovered: %v", r)
		}
		wsh.clientManager.Unregister(client)

		// 設置用戶為離線狀態
		if err := wsh.userService.SetUserOffline(userID); err != nil {
			utils.PrettyPrintf("Failed to set user %s offline: %v", userID, err)
		}

		// 使用互斥鎖保護 WebSocket 連接的關閉操作
		client.WriteMutex.Lock()
		ws.Close()
		client.WriteMutex.Unlock()
	}()

	for {
		var msg WsMessage[json.RawMessage]

		// 設置讀取截止時間
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))

		err := ws.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.PrettyPrintf("WebSocket unexpected close error: %v", err)
			} else {
				utils.PrettyPrintf("Read message failed: %v", err)
			}
			break
		}

		utils.PrettyPrint("Received message:", msg)

		// 重置讀取截止時間
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))

		// 更新用戶活動時間
		if err := wsh.userService.UpdateUserActivity(userID); err != nil {
			utils.PrettyPrintf("Failed to update user activity for %s: %v", userID, err)
		}

		switch msg.Action {
		case "join_room":
			utils.PrettyPrintf("User %s is joining room: %s", userID, msg.Data)
			wsh.handleJoinRoom(ws, client, msg.Data)
		case "leave_room":
			wsh.handleLeaveRoom(ws, client, msg.Data)
		case "send_message":
			utils.PrettyPrintf("User %s is sending message in room: %s", userID, msg.Data)
			wsh.handleSendMessage(ws, client, msg.Data, userID)
		case "ping":
			// 處理客戶端ping
			wsh.handlePing(ws, client)
		default:
			utils.PrettyPrintf("Unknown action: %s", msg.Action)
		}
	}
}

// pingHandler 處理心跳檢測
func (wsh *WebSocketHandler) pingHandler(ws *websocket.Conn, client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		client.WriteMutex.Lock()
		ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
			utils.PrettyPrintf("Ping failed for user %s: %v", client.UserID, err)
			client.WriteMutex.Unlock()
			return
		}
		client.WriteMutex.Unlock()
	}
}

// handlePing 處理ping請求
func (wsh *WebSocketHandler) handlePing(ws *websocket.Conn, client *Client) {
	pongMsg := &WsMessage[map[string]string]{
		Action: "pong",
		Data:   map[string]string{"message": "pong"},
	}

	client.WriteMutex.Lock()
	defer client.WriteMutex.Unlock()

	ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := ws.WriteJSON(pongMsg); err != nil {
		utils.PrettyPrintf("Failed to send pong: %v", err)
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
		client.WriteMutex.Lock()
		ws.WriteJSON(message)
		client.WriteMutex.Unlock()
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
		client.WriteMutex.Lock()
		ws.WriteJSON(message)
		client.WriteMutex.Unlock()
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

	client.WriteMutex.Lock()
	ws.WriteJSON(message)
	client.WriteMutex.Unlock()
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
	client.WriteMutex.Lock()
	ws.WriteJSON(message)
	client.WriteMutex.Unlock()
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

	// 確保房間存在
	wsh.roomManager.InitRoom(requestData.RoomType, requestData.RoomID)

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

	// 使用MessageHandler處理消息
	wsh.messageHandler.HandleMessage(message)
}

// handleDMRoomCreation 處理私聊房間創建邏輯
func (wsh *WebSocketHandler) handleDMRoomCreation(roomID, userID string) {
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

	ctx := context.Background()
	var dmRoomList []models.DMRoom
	err = wsh.odm.Find(ctx, map[string]interface{}{"room_id": roomObjectID}, &dmRoomList)
	if err != nil {
		utils.PrettyPrintf("Failed to find dm room: %v", err)
		return
	}

	var currentUserRoom *models.DMRoom
	var partnerUserRoom *models.DMRoom

	for i := range dmRoomList {
		room := &dmRoomList[i]
		if room.UserID == userObjectID {
			currentUserRoom = room
		} else {
			partnerUserRoom = room
		}
	}

	if currentUserRoom == nil && partnerUserRoom != nil {
		newRoom := &models.DMRoom{
			RoomID:         roomObjectID,
			UserID:         userObjectID,
			ChatWithUserID: partnerUserRoom.UserID,
			IsHidden:       false,
		}
		err := wsh.odm.Create(ctx, newRoom)
		if err != nil {
			utils.PrettyPrintf("Failed to create dm room for user %s: %v", userID, err)
			return
		}
		utils.PrettyPrintf("Created DM room record for user %s in room %s", userID, roomID)
	}

	if partnerUserRoom == nil && currentUserRoom != nil {
		newRoom := &models.DMRoom{
			RoomID:         roomObjectID,
			UserID:         currentUserRoom.ChatWithUserID,
			ChatWithUserID: userObjectID,
			IsHidden:       false,
		}
		err := wsh.odm.Create(ctx, newRoom)
		if err != nil {
			utils.PrettyPrintf("Failed to create dm room for partner: %v", err)
			return
		}
		utils.PrettyPrintf("Created DM room record for partner in room %s", roomID)
	}
}
