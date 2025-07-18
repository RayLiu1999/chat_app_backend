// utils/error_codes.go
package utils

type ErrorCode int

// 錯誤碼常量
const (
	// 通用錯誤
	ErrOperationFailed ErrorCode = 1000 // 操作失敗
	ErrInvalidParams   ErrorCode = 1001 // 請求參數錯誤
	ErrInternalServer  ErrorCode = 1002 // 伺服器內部錯誤

	// 認證錯誤
	ErrUnauthorized ErrorCode = 2000 // 未授權的請求
	ErrLoginFailed  ErrorCode = 2001 // 登入失敗，請檢查信箱或密碼
	ErrLoginExpired ErrorCode = 2002 // 您的登入已過期，請重新登入
	ErrNoPermission ErrorCode = 2003 // 您沒有權限進行此操作
	ErrInvalidToken ErrorCode = 2004 // 無效的 Token

	// 使用者相關錯誤
	ErrUserNotFound   ErrorCode = 3000 // 使用者不存在
	ErrUsernameExists ErrorCode = 3001 // 該用戶名已被使用
	ErrEmailExists    ErrorCode = 3002 // 該信箱已被使用

	// 好友相關錯誤
	ErrFriendExists        ErrorCode = 3500 // 已經是好友
	ErrFriendRequestExists ErrorCode = 3501 // 已經有好友請求

	// 伺服器相關錯誤
	ErrServerNotFound     ErrorCode = 4000 // 伺服器不存在
	ErrNoServerPermission ErrorCode = 4001 // 您沒有權限管理此伺服器
	ErrCreateServerFailed ErrorCode = 4002 // 無法創建伺服器

	// 頻道相關錯誤
	ErrChannelNotFound     ErrorCode = 5000 // 頻道不存在
	ErrCreateChannelFailed ErrorCode = 5001 // 無法創建頻道

	// 訊息相關錯誤
	ErrSendMessageFailed ErrorCode = 6000 // 訊息發送失敗
	ErrGetMessagesFailed ErrorCode = 6001 // 無法獲取訊息歷史記錄

	// 聊天室相關錯誤
	ErrRoomNotFound ErrorCode = 7000 // 聊天室不存在
)

type AppError struct {
	StatusCode  int
	Code        ErrorCode
	Message     string
	Displayable bool
	Err         error // 原始錯誤（可選）
}

// 取得錯誤碼對應的錯誤訊息
// func GetErrorMessage(code ErrorCode) string {
// 	switch code {
// 	// 通用錯誤
// 	case ErrOperationFailed:
// 		return "操作失敗"
// 	case ErrInvalidParams:
// 		return "請求參數錯誤"
// 	case ErrInternalServer:
// 		return "伺服器內部錯誤"

// 	// 認證錯誤
// 	case ErrUnauthorized:
// 		return "未授權的請求"
// 	case ErrLoginFailed:
// 		return "登入失敗，請檢查信箱或密碼"
// 	case ErrLoginExpired:
// 		return "您的登入已過期，請重新登入"
// 	case ErrNoPermission:
// 		return "您沒有權限進行此操作"
// 	case ErrInvalidToken:
// 		return "無效的 Token"

// 	// 使用者相關錯誤
// 	case ErrUserNotFound:
// 		return "使用者不存在"
// 	case ErrUsernameExists:
// 		return "該用戶名已被使用"
// 	case ErrEmailExists:
// 		return "該信箱已被使用"

// 	// 好友相關錯誤
// 	case ErrFriendExists:
// 		return "已經是好友"
// 	case ErrFriendRequestExists:
// 		return "已經有好友請求"

// 	// 伺服器相關錯誤
// 	case ErrServerNotFound:
// 		return "伺服器不存在"
// 	case ErrNoServerPermission:
// 		return "您沒有權限管理此伺服器"
// 	case ErrCreateServerFailed:
// 		return "無法創建伺服器"

// 	// 頻道相關錯誤
// 	case ErrChannelNotFound:
// 		return "頻道不存在"
// 	case ErrCreateChannelFailed:
// 		return "無法創建頻道"

// 	// 訊息相關錯誤
// 	case ErrSendMessageFailed:
// 		return "訊息發送失敗"
// 	case ErrGetMessagesFailed:
// 		return "無法獲取訊息歷史記錄"

// 	default:
// 		return "未知錯誤"
// 	}
// }
