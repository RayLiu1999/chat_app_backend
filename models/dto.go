package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserResponse struct {
	ID         string `json:"id" bson:"_id"`
	Username   string `json:"username" bson:"username"`
	Email      string `json:"email" bson:"email"`
	Nickname   string `json:"nickname" bson:"nickname"`
	PictureURL string `json:"picture_url"`
	BannerURL  string `json:"banner_url"`
}

// LoginResponse 包含登入成功後返回的資訊
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type ServerResponse struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	PictureURL  string             `json:"picture_url" bson:"picture_url"`
	Description string             `json:"description" bson:"description"`
}

// ServerMemberResponse 伺服器成員響應模型
type ServerMemberResponse struct {
	UserID       string `json:"user_id" bson:"user_id"`
	Username     string `json:"username" bson:"username"`
	Nickname     string `json:"nickname" bson:"nickname"` // 伺服器內暱稱
	PictureURL   string `json:"picture_url"`
	Role         string `json:"role" bson:"role"`                     // "owner", "admin", "member"
	IsOnline     bool   `json:"is_online" bson:"is_online"`           // 在線狀態
	LastActiveAt int64  `json:"last_active_at" bson:"last_active_at"` // 最後活動時間
	JoinedAt     int64  `json:"joined_at" bson:"joined_at"`           // 加入時間
}

// ServerDetailResponse 伺服器詳細信息響應（包含成員列表）
type ServerDetailResponse struct {
	ID          primitive.ObjectID     `json:"id" bson:"_id"`
	Name        string                 `json:"name" bson:"name"`
	PictureURL  string                 `json:"picture_url" bson:"picture_url"`
	Description string                 `json:"description" bson:"description"`
	MemberCount int                    `json:"member_count" bson:"member_count"`
	IsPublic    bool                   `json:"is_public" bson:"is_public"`
	OwnerID     string                 `json:"owner_id" bson:"owner_id"`
	Members     []ServerMemberResponse `json:"members" bson:"members"`
	Channels    []ChannelResponse      `json:"channels" bson:"channels"`
}

type ChannelResponse struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	ServerID    primitive.ObjectID `json:"server_id" bson:"server_id"`
	Name        string             `json:"name" bson:"name"`
	Type        string             `json:"type" bson:"type"`
	PictureURL  string             `json:"picture_url" bson:"picture_url"`
	Description string             `json:"description" bson:"description"`
}

type DMRoomResponse struct {
	RoomID     primitive.ObjectID `json:"room_id" bson:"room_id"`
	Nickname   string             `json:"nickname" bson:"nickname"`
	PictureURL string             `json:"picture_url"`
	Timestamp  int64              `json:"timestamp" bson:"timestamp"`
	IsOnline   bool               `json:"is_online" bson:"is_online"` // 聊天對象的在線狀態
}

type MessageResponse struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	RoomType  RoomType           `json:"room_type" bson:"room_type"`
	RoomID    string             `json:"room_id" bson:"room_id"`
	SenderID  string             `json:"sender_id" bson:"sender_id"`
	Content   string             `json:"content" bson:"content"`
	Timestamp int64              `json:"timestamp" bson:"timestamp"`
}

type FriendRequest struct {
	Username string `json:"username" bson:"username"`
}

// FriendResponse 好友響應模型
type FriendResponse struct {
	ID         string `json:"id" bson:"_id"`
	Name       string `json:"name" bson:"name"`
	Nickname   string `json:"nickname" bson:"nickname"`
	PictureURL string `json:"picture_url" bson:"picture_url"`
	Status     string `json:"status" bson:"status"`       // 好友關係狀態：pending, accepted, blocked
	IsOnline   bool   `json:"is_online" bson:"is_online"` // 在線狀態
}

// PendingFriendRequest 待處理好友請求
type PendingFriendRequest struct {
	RequestID  string `json:"request_id" bson:"_id"`
	UserID     string `json:"user_id" bson:"user_id"`
	Username   string `json:"username" bson:"username"`
	Nickname   string `json:"nickname" bson:"nickname"`
	PictureURL string `json:"picture_url" bson:"picture_url"`
	SentAt     int64  `json:"sent_at" bson:"sent_at"`
	Type       string `json:"type" bson:"type"` // "sent" or "received"
}

// PendingRequestsResponse 待處理請求響應
type PendingRequestsResponse struct {
	Sent     []PendingFriendRequest `json:"sent"`     // 我發送的請求
	Received []PendingFriendRequest `json:"received"` // 我收到的請求
	Count    struct {
		Sent     int `json:"sent"`
		Received int `json:"received"`
		Total    int `json:"total"`
	} `json:"count"`
}

// BlockedUserResponse 被封鎖用戶響應
type BlockedUserResponse struct {
	UserID     string `json:"user_id" bson:"user_id"`
	Username   string `json:"username" bson:"username"`
	Nickname   string `json:"nickname" bson:"nickname"`
	PictureURL string `json:"picture_url" bson:"picture_url"`
	BlockedAt  int64  `json:"blocked_at" bson:"blocked_at"`
}

// ServerSearchRequest 伺服器搜尋請求
type ServerSearchRequest struct {
	Query     string `json:"q" form:"q"`                   // 搜尋關鍵字
	Page      int    `json:"page" form:"page"`             // 頁數（從1開始）
	Limit     int    `json:"limit" form:"limit"`           // 每頁數量
	SortBy    string `json:"sort_by" form:"sort_by"`       // 排序方式：name, members, created_at
	SortOrder string `json:"sort_order" form:"sort_order"` // 排序順序：asc, desc
}

// ServerSearchResponse 伺服器搜尋回應
type ServerSearchResponse struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	PictureURL  string             `json:"picture_url" bson:"picture_url"`
	Description string             `json:"description" bson:"description"`
	MemberCount int                `json:"member_count" bson:"member_count"` // 成員數量
	IsJoined    bool               `json:"is_joined" bson:"is_joined"`       // 用戶是否已加入
	OwnerName   string             `json:"owner_name" bson:"owner_name"`     // 伺服器擁有者名稱
	CreatedAt   int64              `json:"created_at" bson:"created_at"`     // 創建時間戳
}

// ServerSearchResults 搜尋結果包裝
type ServerSearchResults struct {
	Servers    []ServerSearchResponse `json:"servers"`
	TotalCount int64                  `json:"total_count"` // 總數量
	Page       int                    `json:"page"`        // 當前頁數
	Limit      int                    `json:"limit"`       // 每頁數量
	TotalPages int                    `json:"total_pages"` // 總頁數
}

// UserProfileResponse 用戶個人資料響應
type UserProfileResponse struct {
	ID         string `json:"id" bson:"_id"`
	Username   string `json:"username" bson:"username"`
	Email      string `json:"email" bson:"email"`
	Nickname   string `json:"nickname" bson:"nickname"`
	PictureURL string `json:"picture_url"` // 圖片 URL（從 PictureID 解析）
	BannerURL  string `json:"banner_url"`  // 橫幅 URL（從 BannerID 解析）
	Status     string `json:"status" bson:"status"`
	Bio        string `json:"bio" bson:"bio"`
}

// UserImageResponse 用戶圖片上傳響應
type UserImageResponse struct {
	ImageURL string `json:"image_url"`
	Type     string `json:"type"` // "avatar" 或 "banner"
}

// TwoFactorStatusResponse 兩步驟驗證狀態響應
type TwoFactorStatusResponse struct {
	Enabled bool `json:"enabled"`
}
