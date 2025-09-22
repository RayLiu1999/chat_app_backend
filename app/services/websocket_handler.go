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
	ws.SetReadLimit(MaxMessageSize)
	ws.SetReadDeadline(time.Now().Add(PongWait))

	// 創建客戶端
	client := wsh.clientManager.NewClient(userID, ws)

	// 設置 pong 處理器
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(PongWait))
		client.UpdateLastSeen()
		return nil
	})

	// 註冊客戶端
	wsh.clientManager.Register(client)

	// 設置用戶為在線狀態
	if err := wsh.userService.SetUserOnline(userID); err != nil {
		utils.PrettyPrintf("無法將用戶 %s 設定為在線：%v", userID, err)
	}

	// 啟動讀寫協程
	go wsh.clientWritePump(client)
	go wsh.clientReadPump(client)

	// 等待客戶端 context 結束
	<-client.Context.Done()

	// 清理工作
	wsh.clientManager.Unregister(client)
	if err := wsh.userService.SetUserOffline(userID); err != nil {
		utils.PrettyPrintf("無法將用戶 %s 設定為離線：%v", userID, err)
	}
}

// handleDMRoomCreation 處理私聊房間創建邏輯
func (wsh *WebSocketHandler) handleDMRoomCreation(roomID, userID string) {
	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		utils.PrettyPrintf("無法解析房間 ID：%v", err)
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.PrettyPrintf("無法解析用戶 ID：%v", err)
		return
	}

	ctx := context.Background()
	var dmRoomList []models.DMRoom
	err = wsh.odm.Find(ctx, map[string]interface{}{"room_id": roomObjectID}, &dmRoomList)
	if err != nil {
		utils.PrettyPrintf("無法找到私聊房間：%v", err)
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
			utils.PrettyPrintf("無法為用戶 %s 創建私聊房間：%v", userID, err)
			return
		}
		utils.PrettyPrintf("已為用戶 %s 在房間 %s 中創建私聊房間記錄", userID, roomID)
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
			utils.PrettyPrintf("無法為對方創建私聊房間：%v", err)
			return
		}
		utils.PrettyPrintf("已為對方在房間 %s 中創建私聊房間記錄", roomID)
	}
}

// clientReadPump 處理客戶端讀取
func (wsh *WebSocketHandler) clientReadPump(client *Client) {
	defer func() {
		if r := recover(); r != nil {
			utils.PrettyPrintf("讀取泵 panic 已恢復，用戶 %s：%v", client.UserID, r)
		}
		client.Cancel() // 取消所有協程
	}()

	for {
		select {
		case <-client.Context.Done():
			return
		default:
			var msg WsMessage[json.RawMessage]

			// 設置讀取超時
			client.Conn.SetReadDeadline(time.Now().Add(PongWait))

			err := client.Conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					utils.PrettyPrintf("WebSocket 意外關閉錯誤，用戶 %s：%v", client.UserID, err)
				} else {
					utils.PrettyPrintf("讀取訊息失敗，用戶 %s：%v", client.UserID, err)
				}
				return
			}

			// 更新最後活動時間
			client.UpdateLastSeen()

			// 更新用戶活動時間
			if err := wsh.userService.UpdateUserActivity(client.UserID); err != nil {
				utils.PrettyPrintf("無法更新用戶 %s 的活動：%v", client.UserID, err)
			}

			// 處理訊息
			wsh.handleClientMessage(client, msg)
		}
	}
}

// clientWritePump 處理客戶端寫入
func (wsh *WebSocketHandler) clientWritePump(client *Client) {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		if r := recover(); r != nil {
			utils.PrettyPrintf("寫入泵 panic 已恢復，用戶 %s：%v", client.UserID, r)
		}
		client.Conn.Close()
	}()

	for {
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
				utils.PrettyPrintf("寫入訊息失敗 for user %s: %v", client.UserID, err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				utils.PrettyPrintf("發送 Ping 失敗 for user %s: %v", client.UserID, err)
				return
			}
		}
	}
}

// handleClientMessage 處理客戶端訊息
func (wsh *WebSocketHandler) handleClientMessage(client *Client, msg WsMessage[json.RawMessage]) {
	switch msg.Action {
	case "join_room":
		utils.PrettyPrintf("用戶 %s 正在加入房間：%s", client.UserID, msg.Data)
		wsh.handleJoinRoom(client, msg.Data)
	case "leave_room":
		wsh.handleLeaveRoom(client, msg.Data)
	case "send_message":
		utils.PrettyPrintf("用戶 %s 正在發送訊息到房間：%s", client.UserID, msg.Data)
		wsh.handleSendMessage(client, msg.Data)
	case "ping":
		// 處理客戶端ping
		wsh.handlePing(client)
	default:
		utils.PrettyPrintf("來自用戶 %s 的未知動作：%s", client.UserID, msg.Action)
		client.SendError("unknown_action", "未知的動作類型")
	}
}

// handleJoinRoom 處理加入房間請求
func (wsh *WebSocketHandler) handleJoinRoom(client *Client, data json.RawMessage) {
	var requestData struct {
		RoomID   string          `json:"room_id"`
		RoomType models.RoomType `json:"room_type"`
	}
	err := json.Unmarshal(data, &requestData)
	if err != nil {
		utils.PrettyPrintf("無法解析加入房間數據：%v", err)
		client.SendError("invalid_data", "無法解析加入房間數據")
		return
	}

	allowed, err := wsh.roomManager.checkUserAllowedJoinRoom(client.UserID, requestData.RoomID, requestData.RoomType)
	if err != nil {
		// 使用統一的 error action
		client.SendError("permission_check_failed", "檢查用戶權限失敗")
		return
	}
	if !allowed {
		// 使用統一的 error action
		client.SendError("permission_denied", "用戶沒有權限加入此房間")
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
func (wsh *WebSocketHandler) handleLeaveRoom(client *Client, data json.RawMessage) {
	var requestData struct {
		RoomID   string          `json:"room_id"`
		RoomType models.RoomType `json:"room_type"`
	}
	err := json.Unmarshal(data, &requestData)
	if err != nil {
		utils.PrettyPrintf("無法解析離開房間數據：%v", err)
		client.SendError("invalid_data", "無法解析離開房間數據")
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
func (wsh *WebSocketHandler) handleSendMessage(client *Client, data json.RawMessage) {
	var requestData struct {
		RoomID   string          `json:"room_id"`
		RoomType models.RoomType `json:"room_type"`
		Content  string          `json:"content"`
	}
	err := json.Unmarshal(data, &requestData)
	if err != nil {
		utils.PrettyPrintf("無法解析發送訊息數據：%v", err)
		client.SendError("invalid_data", "無法解析發送訊息數據")
		return
	}

	// 確保房間存在
	wsh.roomManager.InitRoom(requestData.RoomType, requestData.RoomID)

	// 處理私聊房間邏輯
	if requestData.RoomType == models.RoomTypeDM {
		wsh.handleDMRoomCreation(requestData.RoomID, client.UserID)
	}

	message := &WsMessage[MessageResponse]{
		Data: MessageResponse{
			RoomID:    requestData.RoomID,
			RoomType:  requestData.RoomType,
			SenderID:  client.UserID,
			Content:   requestData.Content,
			Timestamp: time.Now().UnixMilli(),
		},
	}

	// 使用MessageHandler處理消息
	wsh.messageHandler.HandleMessage(message)
}

// handlePing 處理ping請求
func (wsh *WebSocketHandler) handlePing(client *Client) {
	pongMsg := &WsMessage[map[string]interface{}]{
		Action: "pong",
		Data: map[string]interface{}{
			"message":   "pong",
			"timestamp": time.Now().UnixMilli(),
		},
	}

	if err := client.SendMessage(pongMsg); err != nil {
		utils.PrettyPrintf("無法向客戶端 %s 發送 pong：%v", client.UserID, err)
	}
}
