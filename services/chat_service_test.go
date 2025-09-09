package services

import (
	"chat_app_backend/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestChatService_GetUserPictureURL 測試獲取用戶頭像URL
func TestChatService_GetUserPictureURL(t *testing.T) {
	chatService := &ChatService{
		fileUploadService: nil, // 模擬沒有檔案服務
	}

	// 測試沒有頭像ID的情況
	t.Run("NoPictureID", func(t *testing.T) {
		user := &models.User{
			PictureID: primitive.NilObjectID,
		}

		result := chatService.getUserPictureURL(user)
		assert.Empty(t, result)
	})

	// 測試沒有檔案服務的情況
	t.Run("NoFileService", func(t *testing.T) {
		user := &models.User{
			PictureID: primitive.NewObjectID(),
		}

		result := chatService.getUserPictureURL(user)
		assert.Empty(t, result)
	})
}

// TestChatService_CheckUserServerMembership 測試檢查用戶伺服器成員身份
func TestChatService_CheckUserServerMembership(t *testing.T) {
	chatService := &ChatService{}

	// 測試無效的用戶ID
	t.Run("InvalidUserID", func(t *testing.T) {
		isMember, err := chatService.checkUserServerMembership("invalid-id", "507f1f77bcf86cd799439011")
		assert.False(t, isMember)
		assert.NotNil(t, err)
	})

	// 測試無效的伺服器ID
	t.Run("InvalidServerID", func(t *testing.T) {
		isMember, err := chatService.checkUserServerMembership("507f1f77bcf86cd799439011", "invalid-id")
		assert.False(t, isMember)
		assert.NotNil(t, err)
	})
}

// TestRoomKey 測試房間鍵結構
func TestRoomKey(t *testing.T) {
	roomKey := RoomKey{
		Type:   models.RoomTypeDM,
		RoomID: "test-room-123",
	}

	// 測試 String 方法
	expected := "dm:test-room-123"
	actual := roomKey.String()
	assert.Equal(t, expected, actual)

	// 測試不同房間類型
	channelKey := RoomKey{
		Type:   models.RoomTypeChannel,
		RoomID: "channel-456",
	}
	expected = "channel:channel-456"
	actual = channelKey.String()
	assert.Equal(t, expected, actual)
}

// TestClient_SendMessage 測試客戶端發送訊息
func TestClient_SendMessage(t *testing.T) {
	client := &Client{
		UserID: "test-user",
		Send:   make(chan []byte, 10), // 緩衝通道
	}

	// 測試發送成功
	t.Run("SendSuccess", func(t *testing.T) {
		message := map[string]string{"type": "test", "content": "hello"}
		err := client.SendMessage(message)
		assert.Nil(t, err)

		// 檢查訊息是否被發送到通道
		select {
		case msg := <-client.Send:
			assert.Contains(t, string(msg), "test")
			assert.Contains(t, string(msg), "hello")
		case <-time.After(time.Second):
			t.Error("訊息沒有被發送到通道")
		}
	})

	// 測試發送文字訊息
	t.Run("SendText", func(t *testing.T) {
		err := client.SendText("Hello World")
		assert.Nil(t, err)

		// 檢查訊息是否被發送到通道
		select {
		case msg := <-client.Send:
			assert.Equal(t, "Hello World", string(msg))
		case <-time.After(time.Second):
			t.Error("文字訊息沒有被發送到通道")
		}
	})
}

// TestClient_SendError 測試客戶端發送錯誤訊息
func TestClient_SendError(t *testing.T) {
	client := &Client{
		UserID: "test-user",
		Send:   make(chan []byte, 10), // 緩衝通道
	}

	client.SendError("validation_error", "Invalid input")

	// 檢查錯誤訊息是否被發送
	select {
	case msg := <-client.Send:
		assert.Contains(t, string(msg), "error")
		assert.Contains(t, string(msg), "validation_error")
		assert.Contains(t, string(msg), "Invalid input")
	case <-time.After(time.Second):
		t.Error("錯誤訊息沒有被發送到通道")
	}
}

// TestClient_IsHealthy 測試客戶端健康狀態檢查
func TestClient_IsHealthy(t *testing.T) {
	client := &Client{
		UserID:       "test-user",
		IsActive:     true,
		LastPongTime: time.Now(),
	}

	// 測試健康的客戶端
	t.Run("HealthyClient", func(t *testing.T) {
		assert.True(t, client.IsHealthy())
	})

	// 測試非活躍的客戶端
	t.Run("InactiveClient", func(t *testing.T) {
		client.IsActive = false
		assert.False(t, client.IsHealthy())
	})

	// 測試超時的客戶端
	t.Run("TimeoutClient", func(t *testing.T) {
		client.IsActive = true
		client.LastPongTime = time.Now().Add(-PongWait - time.Second) // 超過超時時間
		assert.False(t, client.IsHealthy())
	})
}
