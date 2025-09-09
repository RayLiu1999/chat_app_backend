package models

// APIResponse 定義標準API響應結構
type APIResponse struct {
	Status  string    `json:"status"`            // "success" 或 "error"
	Code    ErrorCode `json:"code,omitempty"`    // 自定義錯誤碼，例如 1001 表示 "用戶不存在"
	Message string    `json:"message,omitempty"` // 訊息內容
	Details any       `json:"details,omitempty"` // 附加的詳細信息
	Data    any       `json:"data,omitempty"`    // 可選的數據
}

// MessageOptions 定義訊息選項
type MessageOptions struct {
	Code    ErrorCode
	Message string
	Details any
}

// ErrorCode 定義錯誤碼類型
type ErrorCode string

// 通用錯誤碼
const (
	// 通用錯誤碼
	ErrOperationFailed ErrorCode = "OPERATION_FAILED" // 操作失敗
	ErrInvalidParams   ErrorCode = "INVALID_PARAMS"   // 請求參數錯誤
	ErrInternalServer  ErrorCode = "INTERNAL_SERVER"  // 伺服器內部錯誤
	ErrNotFound        ErrorCode = "NOT_FOUND"        // 找不到資源
	ErrForbidden       ErrorCode = "FORBIDDEN"        // 禁止訪問
)

// 認證相關錯誤碼
const (
	ErrUnauthorized  ErrorCode = "UNAUTHORIZED"   // 未授權的請求
	ErrLoginFailed   ErrorCode = "LOGIN_FAILED"   // 登入失敗
	ErrLoginExpired  ErrorCode = "LOGIN_EXPIRED"  // 登入已過期
	ErrNoPermission  ErrorCode = "NO_PERMISSION"  // 無權限操作
	ErrInvalidToken  ErrorCode = "INVALID_TOKEN"  // 無效的 Token
	ErrInvalidOrigin ErrorCode = "INVALID_ORIGIN" // 無效的 Origin
)

// 使用者相關錯誤碼
const (
	ErrUserNotFound   ErrorCode = "USER_NOT_FOUND"  // 使用者不存在
	ErrUsernameExists ErrorCode = "USERNAME_EXISTS" // 用戶名已存在
	ErrEmailExists    ErrorCode = "EMAIL_EXISTS"    // 信箱已存在
)

// 好友相關錯誤碼
const (
	ErrFriendExists          ErrorCode = "FRIEND_EXISTS"            // 已是好友
	ErrFriendRequestExists   ErrorCode = "FRIEND_REQUEST_EXISTS"    // 已有好友請求
	ErrFriendRequestNotFound ErrorCode = "FRIEND_REQUEST_NOT_FOUND" // 好友請求不存在
	ErrNotFriends            ErrorCode = "NOT_FRIENDS"              // 不是好友
)

// 伺服器相關錯誤碼
const (
	ErrServerNotFound     ErrorCode = "SERVER_NOT_FOUND"     // 伺服器不存在
	ErrNoServerPermission ErrorCode = "NO_SERVER_PERMISSION" // 無伺服器管理權限
	ErrCreateServerFailed ErrorCode = "CREATE_SERVER_FAILED" // 創建伺服器失敗
)

// 頻道相關錯誤碼
const (
	ErrChannelNotFound     ErrorCode = "CHANNEL_NOT_FOUND"     // 頻道不存在
	ErrCreateChannelFailed ErrorCode = "CREATE_CHANNEL_FAILED" // 創建頻道失敗
)

// 訊息相關錯誤碼
const (
	ErrSendMessageFailed ErrorCode = "SEND_MESSAGE_FAILED" // 訊息發送失敗
	ErrGetMessagesFailed ErrorCode = "GET_MESSAGES_FAILED" // 獲取訊息失敗
)

// 聊天室相關錯誤碼
const (
	ErrRoomNotFound ErrorCode = "ROOM_NOT_FOUND" // 聊天室不存在
)
