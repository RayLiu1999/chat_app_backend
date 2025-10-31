package mocks

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockAuthMiddleware 創建一個模擬認證的中介層，直接設置 user_id
// 這個中介層用於測試環境，繞過真實的 JWT 驗證流程
// 使用方式: router.Use(mocks.MockAuthMiddleware("user123"))
func MockAuthMiddleware(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 直接將 user_id 設置到 context 中
		c.Set("user_id", userID)

		// 設置一個測試用的 ObjectID
		// 如果 userID 是有效的 ObjectID hex string，則使用它
		// 否則使用默認的測試 ObjectID
		objID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			// 使用固定的測試 ObjectID
			objID, _ = primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
		}
		c.Set("user_object_id", objID)

		c.Next()
	}
}

// MockAuthMiddlewareWithObjectID 創建一個模擬認證的中介層，同時指定 userID 和 ObjectID
// 當需要分別控制 userID 字符串和 ObjectID 時使用
func MockAuthMiddlewareWithObjectID(userID string, userObjectID primitive.ObjectID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("user_object_id", userObjectID)
		c.Next()
	}
}
