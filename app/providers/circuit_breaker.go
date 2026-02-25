package providers

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/sony/gobreaker"
)

// newCircuitBreaker 建立一個帶標準設定的 Circuit Breaker
// 設定說明：
//   - MaxRequests: 半開狀態（Half-Open）允許通過的最大探測請求數
//   - Interval: 統計週期，用於計算連續失敗次數
//   - Timeout: 熔斷後多久進入半開狀態嘗試恢復
//   - ReadyToTrip: 觸發熔斷的條件（連續失敗次數 >= 5）
func newCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,                // 半開狀態允許 3 次探測請求
		Interval:    10 * time.Second, // 每 10 秒重置失敗計數
		Timeout:     30 * time.Second, // 熔斷後 30 秒嘗試恢復
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 連續失敗 5 次即觸發熔斷
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			// 狀態變更時記錄日誌，方便監控告警
			slog.Warn("Circuit Breaker 狀態變更",
				"name", name,
				"from", from.String(),
				"to", to.String(),
			)
		},
	}
	return gobreaker.NewCircuitBreaker(settings)
}

// ExecuteWithBreaker 使用熔斷器包裹一個操作
// 如果熔斷器處於開啟（Open）狀態，操作不會執行，直接回傳錯誤
//
// 用法範例：
//
//	result, err := providers.ExecuteWithBreaker(mongoBreaker, func() (interface{}, error) {
//	    return collection.FindOne(ctx, filter)
//	})
func ExecuteWithBreaker(cb *gobreaker.CircuitBreaker, fn func() (interface{}, error)) (interface{}, error) {
	result, err := cb.Execute(fn)
	if err != nil {
		// 特別標記熔斷器自身拋出的錯誤（ErrOpenState 或 ErrTooManyRequests）
		if err == gobreaker.ErrOpenState {
			return nil, fmt.Errorf("服務暫時不可用（熔斷中），請稍後再試: %w", err)
		}
		if err == gobreaker.ErrTooManyRequests {
			return nil, fmt.Errorf("目前請求過多（探測中），請稍後再試: %w", err)
		}
		return nil, err
	}
	return result, nil
}
