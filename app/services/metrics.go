package services

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// WsActiveConnections 當前活躍的 WebSocket 連線數
var WsActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "chat_ws_active_connections",
	Help: "Number of currently active WebSocket connections",
})

// ActiveRooms 當前活躍的房間數（有用戶加入的房間）
var ActiveRooms = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "chat_active_rooms",
	Help: "Number of currently active chat rooms (channel + dm)",
})

// MessagesSavedTotal 已儲存到 DB 的訊息總數（可用 rate() 計算吞吐量）
var MessagesSavedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "chat_messages_saved_total",
	Help: "Total number of chat messages saved to the database",
}, []string{"room_type"})
