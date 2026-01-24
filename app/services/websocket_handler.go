package services

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"chat_app_backend/utils"
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// webSocketHandler 處理 WebSocket 相關操作
type webSocketHandler struct {
	odm            providers.ODM
	clientManager  ClientManager
	roomManager    RoomManager
	messageHandler MessageHandler
	userService    UserService
	cache          providers.CacheProvider
}

// NewWebSocketHandler 創建新的 WebSocket 處理器
func NewWebSocketHandler(odm providers.ODM, clientManager ClientManager, roomManager RoomManager, messageHandler MessageHandler, userService UserService, cache providers.CacheProvider) *webSocketHandler {
	return &webSocketHandler{
		odm:            odm,
		clientManager:  clientManager,
		roomManager:    roomManager,
		messageHandler: messageHandler,
		userService:    userService,
		cache:          cache,
	}
}

// HandleWebSocket 處理 WebSocket 連線
func (wsh *webSocketHandler) HandleWebSocket(ws *websocket.Conn, userID string) {
	// 設置連接參數
	ws.SetReadLimit(MaxMessageSize)
	ws.SetReadDeadline(time.Now().Add(PongWait))

	// 創建客戶端
	client := wsh.clientManager.NewClient(userID, ws)

	// 設置 pong 處理器
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(PongWait))
		if client != nil {
			client.UpdateLastSeen()
		}
		return nil
	})

	// --- 連線建立時的副作用 ---
	// 1. 註冊客戶端到記憶體
	wsh.clientManager.Register(client)

	// 2. 更新資料庫狀態
	if err := wsh.userService.SetUserOnline(userID); err != nil {
		utils.Log.Warn("無法將用戶設定為在線", "user_id", userID, "error", err)
	}

	// 3. 更新 Redis 快取狀態(未來拓展用)
	wsh.cache.Set(utils.UserStatusCacheKey(userID), "online", 24*time.Hour)

	// 啟動讀寫協程
	go wsh.clientWritePump(client)
	go wsh.clientReadPump(client)

	// 等待客戶端 context 結束 (連線關閉)
	if client != nil && client.Context != nil {
		<-client.Context.Done()
	}

	// --- 連線關閉時的清理工作 ---
	// 1. 從記憶體中註銷客戶端
	wsh.clientManager.Unregister(client)

	// 2. 更新資料庫狀態
	if err := wsh.userService.SetUserOffline(userID); err != nil {
		utils.Log.Warn("無法將用戶設定為離線", "user_id", userID, "error", err)
	}

	// 3. 更新 Redis 快取狀態(未來拓展用)
	wsh.cache.Set(utils.UserStatusCacheKey(userID), "offline", 24*time.Hour)
}

// handleDMRoomCreation 處理私聊房間創建邏輯
func (wsh *webSocketHandler) handleDMRoomCreation(roomID, userID string) {
	// 1. 先檢查快取，如果已存在則直接返回 (Hot Path Optimization)
	cacheKey := "dm_room_exists:" + roomID
	if exists, _ := wsh.cache.Get(cacheKey); exists != "" {
		return
	}

	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		utils.Log.Error("無法解析房間 ID", "error", err)
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.Log.Error("無法解析用戶 ID", "error", err)
		return
	}

	ctx := context.Background()
	var dmRoomList []models.DMRoom
	err = wsh.odm.Find(ctx, map[string]any{"room_id": roomObjectID}, &dmRoomList)
	if err != nil {
		utils.Log.Warn("無法找到私聊房間", "room_id", roomID)
		return
	}

	// 如果房間已存在（通常會有兩個 entry，雙方各一個），則寫入快取並返回
	if len(dmRoomList) >= 2 {
		wsh.cache.Set(cacheKey, "1", 24*time.Hour)
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
			utils.Log.Error("無法為用戶創建私聊房間", "user_id", userID, "error", err)
			return
		}
		utils.Log.Debug("已為用戶創建私聊房間記錄", "user_id", userID, "room_id", roomID)
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
			utils.Log.Error("無法為對方創建私聊房間", "error", err)
			return
		}
		utils.Log.Debug("已為對方創建私聊房間記錄", "room_id", roomID)
	}

	// 成功處理後寫入快取
	wsh.cache.Set(cacheKey, "1", 24*time.Hour)
}

// clientReadPump 處理客戶端讀取
func (wsh *webSocketHandler) clientReadPump(client *Client) {
	defer func() {
		if r := recover(); r != nil {
			userID := "unknown"
			if client != nil {
				userID = client.UserID
			}
			utils.Log.Error("讀取泵 panic", "user_id", userID, "panic", r)
		}
		if client != nil {
			client.Cancel() // 取消所有協程
		}
	}()

	for {
		if client == nil || client.Context == nil {
			return
		}
		select {
		case <-client.Context.Done():
			return
		default:
			var msg WsMessage[json.RawMessage]

			// 設置讀取超時
			client.Conn.SetReadDeadline(time.Now().Add(PongWait))

			err := client.Conn.ReadJSON(&msg)
			if err != nil {
				// 只在意外關閉時印日誌，避免大量 EOF/Close 刷屏
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					utils.Log.Warn("WebSocket 意外關閉", "user_id", client.UserID, "error", err)
				} else {
					// 讀取失敗通常意味著斷線，使用 Debug 等級
					utils.Log.Debug("讀取訊息失敗", "user_id", client.UserID, "error", err)
				}
				return
			}

			// 更新最後活動時間 (in-memory)
			client.UpdateLastSeen()

			// 使用 Redis 節流閥更新資料庫中的最後活動時間(未來拓展用)
			wsh.updateActivityWithThrottle(client.UserID)

			// 處理訊息
			wsh.handleClientMessage(client, msg)
		}
	}
}

// updateActivityWithThrottle 使用 Redis 節流閥來更新資料庫中的用戶活動時間
func (wsh *webSocketHandler) updateActivityWithThrottle(userID string) {
	throttleKey := utils.UserActivityThrottleCacheKey(userID)

	// 1. 檢查節流閥是否存在
	val, err := wsh.cache.Get(throttleKey)
	if err != nil {
		// Log a cache error if necessary, but proceed
	}

	// 2. 如果 key 存在，表示在冷卻時間內，直接返回
	if val != "" {
		return
	}

	// 3. 如果 key 不存在，執行更新並設置節流閥
	go func() {
		// 3a. 更新資料庫
		if err := wsh.userService.UpdateUserActivity(userID); err != nil {
			utils.Log.Warn("無法更新用戶活動時間", "user_id", userID, "error", err)
			return // 如果更新失敗，則不設置節流閥，以便下次重試
		}

		// 3b. 設置節流閥，冷卻時間 3 分鐘
		wsh.cache.Set(throttleKey, "1", 3*time.Minute)
	}()
}

// clientWritePump 處理客戶端寫入
func (wsh *webSocketHandler) clientWritePump(client *Client) {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		if r := recover(); r != nil {
			utils.Log.Error("寫入泵 panic", "user_id", client.UserID, "panic", r)
		}
		client.Conn.Close()
	}()

	for {
		if client == nil || client.Context == nil {
			return
		}
		select {
		case <-client.Context.Done():
			return

		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				// 通道已關閉
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				utils.Log.Debug("寫入訊息失敗", "user_id", client.UserID, "error", err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				utils.Log.Debug("發送 Ping 失敗", "user_id", client.UserID, "error", err)
				return
			}
		}
	}
}

// handleClientMessage 處理客戶端訊息
func (wsh *webSocketHandler) handleClientMessage(client *Client, msg WsMessage[json.RawMessage]) {
	switch msg.Action {
	case "join_room":
		wsh.handleJoinRoom(client, msg.Data)
	case "leave_room":
		wsh.handleLeaveRoom(client, msg.Data)
	case "send_message":
		wsh.handleSendMessage(client, msg.Data)
	case "ping":
		// 處理客戶端ping
		wsh.handlePing(client)
	default:
		utils.Log.Warn("未知動作", "user_id", client.UserID, "action", msg.Action)
		client.SendError("unknown_action", "未知的動作類型")
	}
}

// handleJoinRoom 處理加入房間請求
func (wsh *webSocketHandler) handleJoinRoom(client *Client, data json.RawMessage) {
	// 用於錯誤回應的原始動作
	action := "join_room"

	// 解析請求數據
	var requestData struct {
		RoomID   string          `json:"room_id"`
		RoomType models.RoomType `json:"room_type"`
	}
	err := json.Unmarshal(data, &requestData)
	if err != nil {
		utils.Log.Error("無法解析加入房間數據", "error", err)
		client.SendError(action, "無法解析加入房間數據")
		return
	}

	allowed, err := wsh.roomManager.CheckUserAllowedJoinRoom(client.UserID, requestData.RoomID, requestData.RoomType)
	if err != nil {
		// 使用統一的 error action
		client.SendError(action, "檢查用戶權限失敗")
		return
	}
	if !allowed {
		// 使用統一的 error action
		client.SendError(action, "用戶沒有權限加入此房間")
		return
	}

	// 成功時使用 "join_room" action
	wsh.roomManager.InitRoom(requestData.RoomType, requestData.RoomID)
	wsh.roomManager.JoinRoom(client, requestData.RoomType, requestData.RoomID)

	client.SendMessage(&WsMessage[WsStatusResponse]{
		Action: "room_joined",
		Data: WsStatusResponse{
			Status:  "success",
			Message: "成功加入 " + string(requestData.RoomType) + " 房間 " + requestData.RoomID,
		},
	})
}

// handleLeaveRoom 處理離開房間請求
func (wsh *webSocketHandler) handleLeaveRoom(client *Client, data json.RawMessage) {
	// 用於錯誤回應的原始動作
	action := "leave_room"

	// 解析請求數據
	var requestData struct {
		RoomID   string          `json:"room_id"`
		RoomType models.RoomType `json:"room_type"`
	}
	err := json.Unmarshal(data, &requestData)
	if err != nil {
		utils.Log.Error("無法解析離開房間數據", "error", err)
		client.SendError(action, "無法解析離開房間數據")
		return
	}

	wsh.roomManager.LeaveRoom(client, requestData.RoomType, requestData.RoomID)
	client.SendMessage(&WsMessage[WsStatusResponse]{
		Action: "room_left",
		Data: WsStatusResponse{
			Status:  "success",
			Message: "成功離開 " + string(requestData.RoomType) + " 房間 " + requestData.RoomID,
		},
	})
}

// handleSendMessage 處理發送消息請求
func (wsh *webSocketHandler) handleSendMessage(client *Client, data json.RawMessage) {
	// 用於錯誤回應的原始動作
	action := "send_message"

	// 解析請求數據
	var requestData struct {
		RoomID   string          `json:"room_id"`
		RoomType models.RoomType `json:"room_type"`
		Content  string          `json:"content"`
	}
	err := json.Unmarshal(data, &requestData)
	if err != nil {
		utils.Log.Error("無法解析發送訊息數據", "error", err)
		client.SendError(action, "無法解析發送訊息數據")
		return
	}

	// 確保房間存在
	wsh.roomManager.InitRoom(requestData.RoomType, requestData.RoomID)

	// 處理私聊房間邏輯
	if requestData.RoomType == models.RoomTypeDM {
		wsh.handleDMRoomCreation(requestData.RoomID, client.UserID)
	}

	// 建立消息對象
	message := &MessageResponse{
		RoomID:    requestData.RoomID,
		RoomType:  requestData.RoomType,
		SenderID:  client.UserID,
		Content:   requestData.Content,
		Timestamp: time.Now().UnixMilli(),
	}

	// 使用MessageHandler處理消息
	wsh.messageHandler.HandleMessage(message)
}

// handlePing 處理ping請求
func (wsh *webSocketHandler) handlePing(client *Client) {
	pongMsg := &WsMessage[PingResponse]{
		Action: "pong",
		Data: PingResponse{
			Timestamp: time.Now().UnixMilli(),
		},
	}

	if err := client.SendMessage(pongMsg); err != nil {
		client.SendError("ping", "無法發送 pong 訊息")
		utils.Log.Debug("無法向客戶端發送 pong", "user_id", client.UserID, "error", err)
	}
}
