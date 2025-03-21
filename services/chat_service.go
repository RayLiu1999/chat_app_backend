package services

import (
	"chat_app_backend/models"
	"chat_app_backend/repositories"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Client 代表一個WebSocket客戶端
type Client struct {
	ID         string
	UserID     primitive.ObjectID
	Username   string
	Conn       *websocket.Conn
	Send       chan WsMessage // 改為發送 Message 結構體
	ActiveRoom struct {       // 使用者在的聊天室
		ServerID  string // 伺服器
		ChannelID string // 頻道
		RoomID    string // 私聊
	}
	Status    string             // online, offline
	Ctx       context.Context    // 添加context
	CancelCtx context.CancelFunc // 添加取消函數
}

// WsData 定義 WebSocket 數據的接口
type WsData interface {
	GetType() string
}

// Message 聊天消息
type Message struct {
	Type      string `json:"type"` // server, dm
	RoomID    string `json:"room_id"`
	ServerID  string `json:"server_id"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Text      string `json:"text"`
	FileURL   string `json:"file_url"`
	Timestamp int64  `json:"timestamp"`
	Status    string `json:"status"`
}

// GetType 實現 WsData 接口
func (m Message) GetType() string {
	return m.Type
}

// JoinRoom 加入房間的消息
type JoinRoom struct {
	Type     string `json:"type"`
	RoomID   string `json:"room_id"`
	ServerID string `json:"server_id"`
	UserID   string `json:"user_id"`
}

// GetType 實現 WsData 接口
func (j JoinRoom) GetType() string {
	return j.Type
}

// WsMessage WebSocket 消息結構
type WsMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"` // 使用 json.RawMessage 來延遲解析
}

// ChatService 管理所有的聊天功能
type ChatService struct {
	Clients         map[*Client]bool               // 所有連接的客戶端
	ClientsByUserID map[primitive.ObjectID]*Client // 用戶ID到客戶端的映射
	Rooms           map[string]map[*Client]bool    // 房間到客戶端的映射
	Servers         map[string]map[*Client]bool    // 伺服器到客戶端的映射
	Broadcast       chan Message                   // 全局消息通道
	Register        chan *Client                   // 註冊通道
	Unregister      chan *Client                   // 註銷通道
	Lock            sync.Mutex                     // 保護共享資源的鎖
	chatRepo        repositories.ChatRepositoryInterface
	serverRepo      repositories.ServerRepositoryInterface
}

// 用於確保服務只啟動一次的鎖
var runOnce sync.Once

// NewChatService 初始化聊天室服務
func NewChatService(chatRepo repositories.ChatRepositoryInterface, serverRepo repositories.ServerRepositoryInterface) *ChatService {
	cs := &ChatService{
		Clients:         make(map[*Client]bool),
		ClientsByUserID: make(map[primitive.ObjectID]*Client),
		Rooms:           make(map[string]map[*Client]bool),
		Servers:         make(map[string]map[*Client]bool),
		Broadcast:       make(chan Message, 1),  // 增加緩衝，為了避免訊息發送時阻塞
		Register:        make(chan *Client, 10), // 增加緩衝，為了避免使用者重連ws時阻塞
		Unregister:      make(chan *Client, 10), // 增加緩衝，為了避免使用者重連ws時阻塞
		chatRepo:        chatRepo,
		serverRepo:      serverRepo,
	}

	// 確保聊天服務只啟動一次
	runOnce.Do(func() {
		go func() {
			log.Printf("===== ChatService run goroutine 開始啟動 =====")
			cs.run()
			log.Printf("===== ChatService run goroutine 已結束 =====") // 這行正常情況下不應該被執行到
		}()
	})

	return cs
}

// run 處理所有的聊天室邏輯
func (cs *ChatService) run() {
	log.Printf("===== ChatService run() 開始執行 =====")
	count := 0
	for {
		count++
		log.Printf("===== 第 %d 次等待新的事件... =====", count)
		select {
		case client := <-cs.Register:
			log.Printf("===== 處理 Register 事件 =====")
			cs.registerClient(client)

			// 發送上線通知
			cs.updateClientStatus(client, "online")
			log.Printf("===== Register 事件處理完成 =====")

		case client := <-cs.Unregister:
			log.Printf("===== 處理 Unregister 事件 =====")
			cs.unregisterClient(client)

			// 發送離線通知
			cs.updateClientStatus(client, "offline")
			log.Printf("===== Unregister 事件處理完成 =====")

		case message := <-cs.Broadcast:
			log.Printf("===== 處理 Broadcast 事件 =====")
			log.Printf("訊息內容: %+v", message)
			log.Printf("當前所有房間: %+v", cs.Rooms)
			log.Printf("當前所有伺服器: %+v", cs.Servers)

			cs.Lock.Lock() // 在處理廣播時獲取鎖，防止配送過程中客戶端被移除

			switch message.Type {
			case "server":
				log.Printf("處理伺服器訊息，ServerID: %s", message.ServerID)
				if server, ok := cs.Servers[message.ServerID]; ok {
					log.Printf("找到伺服器，當前連接的客戶端數量: %d", len(server))
					for client := range server {
						cs.safelyBroadcastToClient(client, message)
					}
				}
			case "dm":
				log.Printf("處理私聊訊息，RoomID: %s", message.RoomID)
				if room, ok := cs.Rooms[message.RoomID]; ok {
					log.Printf("找到房間，當前連接的客戶端數量: %d", len(room))
					for client := range room {
						cs.safelyBroadcastToClient(client, message)
					}
				}
			}

			cs.Lock.Unlock()
			log.Printf("===== Broadcast 事件處理完成 =====")
		}
	}
}

func (cs *ChatService) safelyBroadcastToClient(client *Client, message Message) {
	// 使用recover處理向已關閉通道發送數據的panic
	defer func() {
		if r := recover(); r != nil {
			log.Printf("向客戶端 %s 發送消息時發生錯誤: %v", client.ID, r)
			// 不在這裡直接調用Unregister，因為我們已經持有鎖
			// 標記這個客戶端稍後需要註銷
			go func() {
				cs.Unregister <- client
			}()
		}
	}()

	var wsMsg WsMessage
	wsMsg.Type = "send_message"

	// 將 Message 編碼為 JSON 並賦值給 wsMsg.Data
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("編碼消息錯誤: %v", err)
		return
	}

	wsMsg.Data = data

	select {
	case client.Send <- wsMsg:
		log.Printf("發送消息到客戶端 %s 成功", client.ID)
		// 更新消息狀態為已送達
		message.Status = "delivered"
	case <-client.Ctx.Done():
		log.Printf("客戶端 %s 已取消", client.ID)
		// 客戶端的上下文已取消，稍後需要註銷
		go func() {
			cs.Unregister <- client
		}()
	default:
		log.Printf("客戶端 %s 的通道已滿或阻塞", client.ID)
		// 通道已滿或阻塞，需要註銷客戶端
		go func() {
			cs.Unregister <- client
		}()
	}
}

// HandleConnection 處理新的WebSocket連接
func (cs *ChatService) HandleConnection(userID primitive.ObjectID, conn *websocket.Conn) {
	log.Print("Handling connection...")

	// 創建帶取消功能的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 創建新連接
	client := &Client{
		ID:     uuid.New().String(),
		UserID: userID,
		Conn:   conn,
		Send:   make(chan WsMessage, 256),
		Status: "online",
		ActiveRoom: struct {
			ServerID  string
			ChannelID string
			RoomID    string
		}{
			RoomID: "fdssd",
		},
		Ctx:       ctx,
		CancelCtx: cancel,
	}

	log.Printf("===== 新客戶端連接:%v =====", client)

	// 註冊新客戶端
	cs.Register <- client
	log.Printf("===== 註冊請求已發送 =====")

	// 啟動讀寫協程
	go cs.readPump(client)
	go cs.writePump(client)
}

// readPump 處理來自客戶端的消息
func (cs *ChatService) readPump(client *Client) {
	defer func() {
		client.CancelCtx() // 取消context
		cs.Unregister <- client
		client.Conn.Close()
	}()

	for {
		var wsMsg WsMessage
		err := client.Conn.ReadJSON(&wsMsg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("收到新消息：%+v", wsMsg)

		// 根據消息類型解析數據
		switch wsMsg.Type {
		case "send_message":
			var msg Message
			if err := json.Unmarshal(wsMsg.Data, &msg); err != nil {
				log.Printf("解析聊天消息錯誤: %v", err)
				continue
			}

			// 處理傳送訊息的邏輯
			log.Printf("===== readPump 收到新訊息 =====")
			log.Printf("訊息內容: %+v", msg)
			log.Printf("userId: %s", msg.UserID)

			switch msg.Type {
			case "dm":
				// msg.UserID = client.ID
				// msg.Username = client.Username
				// msg.Timestamp = time.Now().Unix()
				msg.Status = "sent" // 設定為 "sent"
				// msg.RoomID = client.ActiveRoom.RoomID

				log.Printf("===== 準備發送訊息到 Broadcast 通道 =====")
				select {
				case cs.Broadcast <- msg:
					log.Printf("===== 訊息已成功發送到 Broadcast 通道 =====")
				default:
					log.Printf("===== 警告：Broadcast 通道阻塞，訊息無法發送 =====")
				}
			}

		case "join_room":
			var joinRoom JoinRoom
			if err := json.Unmarshal(wsMsg.Data, &joinRoom); err != nil {
				log.Printf("解析加入房間消息錯誤: %v", err)
				continue
			}
			log.Printf("收到加入房間請求: %+v", joinRoom)
			// 處理加入房間...

		default:
			log.Printf("未知的消息類型: %s", wsMsg.Type)
		}
	}
}

// writePump 將消息發送給客戶端
func (cs *ChatService) writePump(client *Client) {
	defer client.Conn.Close()

	for {
		select {
		case <-client.Ctx.Done():
			log.Printf("writePump for client %s terminated by context", client.ID)
			return
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Printf("Sending message: %v", message)

			// 使用 WriteJSON 發送消息
			if err := client.Conn.WriteJSON(message); err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}
		}
	}
}

// 寫入聊天資料表記錄
func (cs *ChatService) writeChatRecord(message Message) {
	// 將 WebSocket 消息轉換為數據庫模型
	dbMessage := models.Message{
		ID:        primitive.NewObjectID(),
		Type:      message.Type,
		Content:   message.Text,
		CreatedAt: time.Now(),
		UpdateAt:  time.Now(),
	}

	// 轉換 UserID 為 ObjectID
	senderID, err := primitive.ObjectIDFromHex(message.UserID)
	if err != nil {
		log.Printf("無效的用戶ID: %v", err)
		return
	}
	dbMessage.SenderID = senderID

	// 轉換 RoomID 為 ObjectID
	roomID, err := primitive.ObjectIDFromHex(message.RoomID)
	if err != nil {
		log.Printf("無效的房間ID: %v", err)
		return
	}
	dbMessage.RoomID = roomID

	// 如果是伺服器消息，則不設置接收者ID
	if message.Type == "dm" && message.RoomID != "" {
		// 對於私聊消息，我們需要找出接收者ID
		// 這裡假設私聊房間只有兩個人
		cs.Lock.Lock()
		if room, ok := cs.Rooms[message.RoomID]; ok {
			for client := range room {
				if client.UserID.Hex() != message.UserID {
					dbMessage.ReceiverID = client.UserID
					break
				}
			}
		}
		cs.Lock.Unlock()
	}

	// 保存消息到數據庫
	_, err = cs.chatRepo.SaveMessage(dbMessage)
	if err != nil {
		log.Printf("保存聊天記錄失敗: %v", err)
	} else {
		log.Printf("聊天記錄已保存到數據庫, 消息ID: %s", dbMessage.ID.Hex())
	}
}

// updateClientStatus 更新客戶端狀態
func (cs *ChatService) updateClientStatus(client *Client, status string) {
	cs.Lock.Lock()
	client.Status = status
	cs.Lock.Unlock()
}

func (cs *ChatService) registerClient(client *Client) {
	cs.Lock.Lock()
	defer cs.Lock.Unlock()

	// 清理舊連接
	if oldClient, exists := cs.ClientsByUserID[client.UserID]; exists {
		cs.unregisterClientUnlocked(oldClient)
	}

	// 註冊新連接
	cs.Clients[client] = true
	cs.ClientsByUserID[client.UserID] = client

	// 加入房間
	if client.ActiveRoom.RoomID != "" {
		if _, ok := cs.Rooms[client.ActiveRoom.RoomID]; !ok {
			cs.Rooms[client.ActiveRoom.RoomID] = make(map[*Client]bool)
		}
		cs.Rooms[client.ActiveRoom.RoomID][client] = true
	}

	// 加入伺服器
	if client.ActiveRoom.ServerID != "" {
		if _, ok := cs.Servers[client.ActiveRoom.ServerID]; !ok {
			cs.Servers[client.ActiveRoom.ServerID] = make(map[*Client]bool)
		}
		cs.Servers[client.ActiveRoom.ServerID][client] = true
	}
}

func (cs *ChatService) unregisterClientUnlocked(client *Client) {
	// 這個版本不獲取鎖，呼叫者負責鎖管理
	if _, ok := cs.Clients[client]; ok {
		delete(cs.Clients, client)
		delete(cs.ClientsByUserID, client.UserID)

		// 從房間中移除
		if client.ActiveRoom.RoomID != "" {
			if room, ok := cs.Rooms[client.ActiveRoom.RoomID]; ok {
				delete(room, client)
				if len(room) == 0 {
					delete(cs.Rooms, client.ActiveRoom.RoomID)
				}
			}
		}

		// 從伺服器中移除
		if client.ActiveRoom.ServerID != "" {
			if server, ok := cs.Servers[client.ActiveRoom.ServerID]; ok {
				delete(server, client)
				if len(server) == 0 {
					delete(cs.Servers, client.ActiveRoom.ServerID)
				}
			}
		}

		// 嘗試安全關閉通道
		client.CancelCtx() // 確保所有協程都能收到取消信號
		close(client.Send)
	}
}

func (cs *ChatService) unregisterClient(client *Client) {
	cs.Lock.Lock()
	defer cs.Lock.Unlock()
	cs.unregisterClientUnlocked(client)
}

// 添加一個輔助函數來處理連接關閉
func (cs *ChatService) handleConnectionClose(client *Client, code int, text string) {
	log.Printf("WebSocket 連接關閉事件 - 客戶端: %s, 代碼: %d, 原因: %s", client.ID, code, text)
	switch code {
	case websocket.CloseGoingAway:
		log.Printf("客戶端 %s 正常離開（頁面關閉/刷新）", client.ID)
	case websocket.CloseAbnormalClosure:
		log.Printf("客戶端 %s 異常關閉", client.ID)
	case websocket.CloseNoStatusReceived:
		log.Printf("客戶端 %s 沒有發送關閉狀態就斷開了", client.ID)
	default:
		log.Printf("客戶端 %s 其他關閉情況: %d", client.ID, code)
	}
}

// GetServerListByUserId 獲取用戶的伺服器列表
func (cs *ChatService) GetServerListByUserId(objectID primitive.ObjectID) ([]models.Server, error) {
	servers, err := cs.serverRepo.GetServerListByUserId(objectID)
	if err != nil {
		return nil, err
	}

	return servers, nil
}

// AddServer 添加新伺服器
func (cs *ChatService) AddServer(server *models.Server) (models.Server, error) {
	// ... 保持原有的實現 ...
	return models.Server{}, nil
}
