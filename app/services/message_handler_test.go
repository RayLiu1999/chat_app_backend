package services

import (
	"chat_app_backend/app/models"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// TestNewMessageHandler 測試創建消息處理器
func TestNewMessageHandler(t *testing.T) {
	// 由於 NewMessageHandler 需要 ODM 和 RoomManager，我們傳入 nil 進行基本測試
	handler := NewMessageHandler(nil, nil, nil)

	assert.NotNil(t, handler)
}

// TestIsClientConnectionValid 測試檢查客戶端連線有效性
func TestIsClientConnectionValid(t *testing.T) {
	mockConn := &websocket.Conn{}
	handler := NewMessageHandler(nil, nil, nil)

	tests := []struct {
		name     string
		client   *Client
		expected bool
	}{
		{
			name:     "客戶端為 nil",
			client:   nil,
			expected: false,
		},
		{
			name: "連線為 nil",
			client: &Client{
				UserID:       "user1",
				Conn:         nil,
				IsActive:     true,
				LastPongTime: time.Now(),
				ConnectedAt:  time.Now(),
			},
			expected: false,
		},
		{
			name: "客戶端已標記為非活躍",
			client: &Client{
				UserID:       "user1",
				Conn:         mockConn,
				IsActive:     false,
				LastPongTime: time.Now(),
				ConnectedAt:  time.Now(),
			},
			expected: false,
		},
		{
			name: "最後 Pong 時間過久",
			client: &Client{
				UserID:       "user1",
				Conn:         mockConn,
				IsActive:     true,
				LastPongTime: time.Now().Add(-3 * time.Minute),
				ConnectedAt:  time.Now(),
			},
			expected: false,
		},
		{
			name: "連線時間過久",
			client: &Client{
				UserID:       "user1",
				Conn:         mockConn,
				IsActive:     true,
				LastPongTime: time.Now(),
				ConnectedAt:  time.Now().Add(-25 * time.Hour),
			},
			expected: false,
		},
		{
			name: "有效的客戶端連線",
			client: &Client{
				UserID:       "user1",
				Conn:         mockConn,
				IsActive:     true,
				LastPongTime: time.Now(),
				ConnectedAt:  time.Now(),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.isClientConnectionValid(tt.client)
			assert.Equal(t, tt.expected, result)

			// 對於時間過久的情況，驗證 IsActive 被設置為 false
			if tt.client != nil && !tt.expected && tt.name != "客戶端為 nil" && tt.name != "連線為 nil" && tt.name != "客戶端已標記為非活躍" {
				if time.Since(tt.client.LastPongTime) > 2*time.Minute || time.Since(tt.client.ConnectedAt) > 24*time.Hour {
					assert.False(t, tt.client.IsActive, "客戶端應該被標記為非活躍")
				}
			}
		})
	}
}

// TestIsClientConnectionValid_WithRealTime 測試實際時間場景
func TestIsClientConnectionValid_WithRealTime(t *testing.T) {
	mockConn := &websocket.Conn{}
	handler := NewMessageHandler(nil, nil, nil)

	t.Run("剛連線的客戶端應該有效", func(t *testing.T) {
		client := &Client{
			UserID:       "user1",
			Conn:         mockConn,
			IsActive:     true,
			LastPongTime: time.Now(),
			ConnectedAt:  time.Now(),
		}

		result := handler.isClientConnectionValid(client)
		assert.True(t, result)
		assert.True(t, client.IsActive)
	})

	t.Run("Pong 超時應該失敗並標記非活躍", func(t *testing.T) {
		client := &Client{
			UserID:       "user1",
			Conn:         mockConn,
			IsActive:     true,
			LastPongTime: time.Now().Add(-2*time.Minute - 1*time.Second),
			ConnectedAt:  time.Now(),
		}

		result := handler.isClientConnectionValid(client)
		assert.False(t, result)
		assert.False(t, client.IsActive)
	})

	t.Run("連線超過 24 小時應該失敗並標記非活躍", func(t *testing.T) {
		client := &Client{
			UserID:       "user1",
			Conn:         mockConn,
			IsActive:     true,
			LastPongTime: time.Now(),
			ConnectedAt:  time.Now().Add(-24*time.Hour - 1*time.Second),
		}

		result := handler.isClientConnectionValid(client)
		assert.False(t, result)
		assert.False(t, client.IsActive)
	})
}

// TestMessageResponse_Structure 測試 MessageResponse 結構
func TestMessageResponse_Structure(t *testing.T) {
	t.Run("創建有效的 MessageResponse", func(t *testing.T) {
		timestamp := time.Now().Unix()
		msg := &MessageResponse{
			RoomID:    "room123",
			SenderID:  "user123",
			Content:   "Hello World",
			RoomType:  models.RoomTypeChannel,
			Timestamp: timestamp,
		}

		assert.Equal(t, "room123", msg.RoomID)
		assert.Equal(t, "user123", msg.SenderID)
		assert.Equal(t, "Hello World", msg.Content)
		assert.Equal(t, models.RoomTypeChannel, msg.RoomType)
		assert.NotZero(t, msg.Timestamp)
	})

	t.Run("測試不同的房間類型", func(t *testing.T) {
		channelMsg := &MessageResponse{
			RoomType: models.RoomTypeChannel,
		}
		assert.Equal(t, models.RoomTypeChannel, channelMsg.RoomType)

		dmMsg := &MessageResponse{
			RoomType: models.RoomTypeDM,
		}
		assert.Equal(t, models.RoomTypeDM, dmMsg.RoomType)
	})
}
