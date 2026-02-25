package middlewares

import (
	"context"
	"net/http"
	"time"

	"chat_app_backend/app/models"

	"github.com/gin-gonic/gin"
)

// Timeout 為每個 HTTP 請求設定最長處理時間
// 超過時間後，context 會被取消，DB 查詢等依賴 context 的操作將中止
func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 建立帶有 timeout 的 context
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()

		// 將新 context 注入請求，讓下游的 repository 呼叫可以繼承 timeout
		c.Request = c.Request.WithContext(ctx)

		// 使用 channel 等待 handler 完成或 context 超時
		finished := make(chan struct{}, 1)

		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// 正常完成，不需額外處理
		case <-ctx.Done():
			// Context 超時，回傳 504 錯誤
			// 注意：此時 c.Next() 可能仍在執行，但 DB 查詢因 ctx 取消而中止
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, models.APIResponse{
				Status:  "error",
				Code:    models.ErrInternalServer,
				Message: "請求處理超時，請稍後再試",
			})
		}
	}
}
