package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"chat_app_backend/repositories"
	"chat_app_backend/utils"
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	config            *config.Config
	userRepo          repositories.UserRepositoryInterface
	odm               *providers.ODM
	clientManager     *ClientManager             // 添加 ClientManager 依賴
	fileUploadService FileUploadServiceInterface // 添加 FileUploadService 依賴
}

func NewUserService(cfg *config.Config, odm *providers.ODM, userRepo repositories.UserRepositoryInterface, clientManager *ClientManager, fileUploadService FileUploadServiceInterface) *UserService {
	return &UserService{
		config:            cfg,
		userRepo:          userRepo,
		odm:               odm,
		clientManager:     clientManager,
		fileUploadService: fileUploadService,
	}
}

// getUserPictureURL 獲取用戶頭像 URL（從 ObjectID 解析）
func (us *UserService) getUserPictureURL(user *models.User) string {
	if user.PictureID.IsZero() || us.fileUploadService == nil {
		return ""
	}

	pictureURL, err := us.fileUploadService.GetFileURLByID(user.PictureID.Hex())
	if err != nil {
		return ""
	}
	return pictureURL
}

// getUserBannerURL 獲取用戶橫幅 URL（從 ObjectID 解析）
func (us *UserService) getUserBannerURL(user *models.User) string {
	if user.BannerID.IsZero() || us.fileUploadService == nil {
		return ""
	}

	bannerURL, err := us.fileUploadService.GetFileURLByID(user.BannerID.Hex())
	if err != nil {
		return ""
	}
	return bannerURL
}

// 註冊新用戶
func (us *UserService) RegisterUser(user models.User) *utils.AppError {
	// 檢查用戶名是否已存在
	exists, err := us.userRepo.CheckUsernameExists(user.Username)
	if err != nil {
		return &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}
	if exists {
		return &utils.AppError{
			Code: utils.ErrUsernameExists,
		}
	}

	// 檢查電子郵件是否已存在
	exists, err = us.userRepo.CheckEmailExists(user.Email)
	if err != nil {
		return &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}
	if exists {
		return &utils.AppError{
			Code: utils.ErrEmailExists,
		}
	}

	// 加密密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}
	user.Password = string(hashedPassword)

	// 設置創建時間和更新時間
	now := time.Now()
	user.BaseModel.CreatedAt = now
	user.BaseModel.UpdatedAt = now

	// 創建用戶
	err = us.userRepo.CreateUser(user)
	if err != nil {
		return &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}

	return nil
}

// 根據ID獲取用戶信息
func (us *UserService) GetUserResponseById(userID string) (*models.UserResponse, error) {
	user, err := us.userRepo.GetUserById(userID)
	if err != nil {
		return nil, err
	}

	// 轉換為 UserResponse
	response := &models.UserResponse{
		ID:         userID,
		Username:   user.Username,
		Email:      user.Email,
		Nickname:   user.Nickname,
		PictureURL: us.getUserPictureURL(user),
		BannerURL:  us.getUserBannerURL(user),
	}

	return response, nil
}

// Login 處理用戶登入邏輯
func (us *UserService) Login(loginUser models.User) (*models.LoginResponse, *utils.AppError) {
	// 刪除過期或被註銷的 refresh token
	err := us.ClearExpiredRefreshTokens()
	if err != nil {
		return nil, &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}

	// 查找用戶
	var user models.User
	err = us.odm.FindOne(context.Background(), bson.M{"email": loginUser.Email}, &user)
	if err != nil {
		return nil, &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}

	// 驗證密碼
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginUser.Password))
	if err != nil {
		return nil, &utils.AppError{
			Code:        utils.ErrLoginFailed,
			Displayable: false,
		}
	}

	// 生成 Refresh Token
	refreshTokenResponse, err := utils.GenRefreshToken(user.BaseModel.GetID().Hex())
	if err != nil {
		return nil, &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}

	// 將 refresh token 寫入資料庫
	var refreshTokenDoc = models.RefreshToken{
		UserID:    user.BaseModel.GetID(),
		Token:     refreshTokenResponse.Token,
		ExpiresAt: refreshTokenResponse.ExpiresAt,
		Revoked:   false,
	}

	err = us.odm.Create(context.Background(), &refreshTokenDoc)
	if err != nil {
		return nil, &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}

	// 生成 Access Token
	accessTokenResponse, err := utils.GenAccessToken(user.BaseModel.GetID().Hex())
	if err != nil {
		return nil, &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}

	// 返回 tokens
	return &models.LoginResponse{
		AccessToken:  accessTokenResponse.Token,
		RefreshToken: refreshTokenResponse.Token,
	}, nil
}

// 登出
func (us *UserService) Logout(c *gin.Context) *utils.AppError {
	// 註銷 refresh token
	_, userObjectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		return &utils.AppError{
			Code: utils.ErrUnauthorized,
		}
	}

	// 使用 UpdateMany 直接更新所有符合條件的文檔
	filter := bson.M{"user_id": userObjectID, "revoked": false}
	update := bson.M{"$set": bson.M{"revoked": true}}
	err = us.odm.UpdateMany(context.Background(), &models.RefreshToken{}, filter, update)
	if err != nil {
		return &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}

	// 清除 cookie
	c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	return nil
}

// RefreshToken 刷新令牌
func (us *UserService) RefreshToken(refreshToken string) (string, *utils.AppError) {
	// 查詢 refresh token
	var refreshTokenDoc models.RefreshToken
	err := us.odm.FindOne(context.Background(), bson.M{"token": refreshToken}, &refreshTokenDoc)
	if err != nil {
		return "", &utils.AppError{
			Code:        utils.ErrInvalidToken,
			Displayable: false,
		}
	}

	// 檢查是否被註銷或過期
	if refreshTokenDoc.Revoked || refreshTokenDoc.ExpiresAt < time.Now().Unix() {
		// 移除 refresh token
		err = us.odm.Delete(context.Background(), &refreshTokenDoc)
		if err != nil {
			return "", &utils.AppError{
				Code: utils.ErrInternalServer,
				Err:  err,
			}
		}

		return "", &utils.AppError{
			Code:        utils.ErrInvalidToken,
			Displayable: false,
		}
	}

	// 生成新的 access token
	accessTokenResponse, err := utils.GenAccessToken(refreshTokenDoc.UserID.Hex())
	if err != nil {
		return "", &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}

	// 更新 refresh token 與過期時間
	refreshTokenDoc.ExpiresAt = time.Now().Add(time.Hour * 24 * 7).Unix()
	err = us.odm.Update(context.Background(), &refreshTokenDoc)
	if err != nil {
		return "", &utils.AppError{
			Code: utils.ErrInternalServer,
			Err:  err,
		}
	}

	return accessTokenResponse.Token, nil
}

// 清除過期或被註銷的 refresh token
func (us *UserService) ClearExpiredRefreshTokens() error {
	filter := bson.M{"$or": []bson.M{
		{"expires_at": bson.M{"$lt": time.Now().Unix()}},
		{"revoked": true},
	}}
	err := us.odm.DeleteMany(context.Background(), &models.RefreshToken{}, filter)
	return err
}

// SetUserOnline 設置用戶為在線狀態
func (us *UserService) SetUserOnline(userID string) error {
	return us.userRepo.UpdateUserOnlineStatus(userID, true)
}

// SetUserOffline 設置用戶為離線狀態
func (us *UserService) SetUserOffline(userID string) error {
	return us.userRepo.UpdateUserOnlineStatus(userID, false)
}

// UpdateUserActivity 更新用戶活動時間（保持在線狀態）
func (us *UserService) UpdateUserActivity(userID string) error {
	timestamp := time.Now().Unix()
	return us.userRepo.UpdateUserLastActiveTime(userID, timestamp)
}

// IsUserOnlineByWebSocket 基於 WebSocket 連線檢查用戶是否在線
func (us *UserService) IsUserOnlineByWebSocket(userID string) bool {
	_, exists := us.clientManager.GetClient(userID)
	return exists
}

// CheckAndSetOfflineUsers 檢查並設置離線用戶（定期任務用）
// 現在這個方法主要用於數據庫狀態同步，實際在線狀態以 WebSocket 為準
func (us *UserService) CheckAndSetOfflineUsers(offlineThresholdMinutes int) error {
	// 計算離線閾值時間戳
	thresholdTimestamp := time.Now().Add(-time.Duration(offlineThresholdMinutes) * time.Minute).Unix()

	// 查找超過閾值時間未活動的在線用戶
	filter := bson.M{
		"is_online":      true,
		"last_active_at": bson.M{"$lt": thresholdTimestamp},
	}

	update := bson.M{
		"$set": bson.M{
			"is_online":  false,
			"updated_at": time.Now(),
		},
	}

	return us.odm.UpdateMany(context.Background(), &models.User{}, filter, update)
}

// GetUserProfile 獲取用戶個人資料
func (us *UserService) GetUserProfile(userID string) (*models.UserProfileResponse, error) {
	user, err := us.userRepo.GetUserById(userID)
	if err != nil {
		return nil, err
	}

	profile := &models.UserProfileResponse{
		ID:         user.ID.Hex(),
		Username:   user.Username,
		Email:      user.Email,
		Nickname:   user.Nickname,
		PictureURL: us.getUserPictureURL(user),
		BannerURL:  us.getUserBannerURL(user),
		Status:     user.Status,
		Bio:        user.Bio,
	}

	// 解析圖片 URL
	if !user.PictureID.IsZero() {
		if pictureURL, err := us.fileUploadService.GetFileURLByID(user.PictureID.Hex()); err == nil && pictureURL != "" {
			profile.PictureURL = pictureURL
		}
	}

	if !user.BannerID.IsZero() {
		if bannerURL, err := us.fileUploadService.GetFileURLByID(user.BannerID.Hex()); err == nil && bannerURL != "" {
			profile.BannerURL = bannerURL
		}
	}

	return profile, nil
}

// UpdateUserProfile 更新用戶基本資料
func (us *UserService) UpdateUserProfile(userID string, updates map[string]interface{}) error {
	// 過濾允許更新的欄位
	allowedFields := map[string]bool{
		"username": true,
		"nickname": true,
		"status":   true,
		"bio":      true,
	}

	filteredUpdates := make(map[string]interface{})
	for field, value := range updates {
		if allowedFields[field] {
			filteredUpdates[field] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return nil // 沒有需要更新的欄位
	}

	// 添加更新時間
	filteredUpdates["updated_at"] = time.Now()

	return us.userRepo.UpdateUser(userID, filteredUpdates)
}

// UploadUserImage 上傳用戶頭像或橫幅
func (us *UserService) UploadUserImage(userID string, file multipart.File, header *multipart.FileHeader, imageType string) (*models.UserImageResponse, error) {
	if us.fileUploadService == nil {
		return nil, fmt.Errorf("檔案上傳服務未初始化")
	}

	var config *models.FileUploadConfig
	var fieldName string

	// 根據圖片類型選擇配置
	switch imageType {
	case "avatar":
		config = models.GetAvatarUploadConfig()
		fieldName = "picture_id"
	case "banner":
		config = models.GetBannerUploadConfig()
		fieldName = "banner_id"
	default:
		return nil, fmt.Errorf("不支援的圖片類型: %s", imageType)
	}

	// 上傳檔案
	uploadResult, err := us.fileUploadService.UploadFileWithConfig(file, header, userID, config)
	if err != nil {
		return nil, fmt.Errorf("圖片上傳失敗: %v", err)
	}

	// 獲取圖片URL
	imageURL, err := us.fileUploadService.GetFileURLByID(uploadResult.ID.Hex())
	if err != nil {
		// 如果獲取URL失敗，嘗試刪除已上傳的檔案
		if deleteErr := us.fileUploadService.DeleteFileByID(uploadResult.ID.Hex(), userID); deleteErr != nil {
			fmt.Printf("清理上傳檔案失敗: %v\n", deleteErr)
		}
		return nil, fmt.Errorf("獲取圖片URL失敗: %v", err)
	}

	// 更新用戶資料庫記錄（儲存檔案ID）
	updates := map[string]interface{}{
		fieldName:    uploadResult.ID,
		"updated_at": time.Now(),
	}

	err = us.userRepo.UpdateUser(userID, updates)
	if err != nil {
		// 如果更新資料庫失敗，嘗試刪除已上傳的檔案
		if deleteErr := us.fileUploadService.DeleteFileByID(uploadResult.ID.Hex(), userID); deleteErr != nil {
			fmt.Printf("清理上傳檔案失敗: %v\n", deleteErr)
		}
		return nil, fmt.Errorf("更新用戶資料失敗: %v", err)
	}

	return &models.UserImageResponse{
		ImageURL: imageURL,
		Type:     imageType,
	}, nil
}

// DeleteUserAvatar 刪除用戶頭像
func (us *UserService) DeleteUserAvatar(userID string) error {
	// 獲取用戶當前資料，以便刪除舊的頭像檔案
	user, err := us.userRepo.GetUserById(userID)
	if err != nil {
		return err
	}

	// 如果有舊的頭像，嘗試刪除檔案
	if !user.PictureID.IsZero() && us.fileUploadService != nil {
		if err := us.fileUploadService.DeleteFileByID(user.PictureID.Hex(), userID); err != nil {
			// 記錄錯誤但不阻止更新資料庫
			fmt.Printf("刪除頭像檔案失敗: %v\n", err)
		}
	}

	// 更新資料庫記錄，清空圖片ID
	updates := map[string]interface{}{
		"picture_id": nil,
		"updated_at": time.Now(),
	}
	return us.userRepo.UpdateUser(userID, updates)
}

// DeleteUserBanner 刪除用戶橫幅
func (us *UserService) DeleteUserBanner(userID string) error {
	// 獲取用戶當前資料，以便刪除舊的橫幅檔案
	user, err := us.userRepo.GetUserById(userID)
	if err != nil {
		return err
	}

	// 如果有舊的橫幅，嘗試刪除檔案
	if !user.BannerID.IsZero() && us.fileUploadService != nil {
		if err := us.fileUploadService.DeleteFileByID(user.BannerID.Hex(), userID); err != nil {
			// 記錄錯誤但不阻止更新資料庫
			fmt.Printf("刪除橫幅檔案失敗: %v\n", err)
		}
	}

	// 更新資料庫記錄，清空圖片ID
	updates := map[string]interface{}{
		"banner_id":  nil,
		"updated_at": time.Now(),
	}
	return us.userRepo.UpdateUser(userID, updates)
}

// UpdateUserPassword 更新用戶密碼
func (us *UserService) UpdateUserPassword(userID string, newPassword string) error {
	// 對新密碼進行雜湊
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"password":   string(hashedPassword),
		"updated_at": time.Now(),
	}

	return us.userRepo.UpdateUser(userID, updates)
}

// GetTwoFactorStatus 獲取兩步驟驗證狀態
func (us *UserService) GetTwoFactorStatus(userID string) (*models.TwoFactorStatusResponse, error) {
	user, err := us.userRepo.GetUserById(userID)
	if err != nil {
		return nil, err
	}

	return &models.TwoFactorStatusResponse{
		Enabled: user.TwoFactorEnabled,
	}, nil
}

// UpdateTwoFactorStatus 啟用/停用兩步驟驗證
func (us *UserService) UpdateTwoFactorStatus(userID string, enabled bool) error {
	updates := map[string]interface{}{
		"two_factor_enabled": enabled,
		"updated_at":         time.Now(),
	}

	return us.userRepo.UpdateUser(userID, updates)
}

// DeactivateAccount 停用帳號
func (us *UserService) DeactivateAccount(userID string) error {
	updates := map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	}

	return us.userRepo.UpdateUser(userID, updates)
}

// DeleteAccount 刪除帳號
func (us *UserService) DeleteAccount(userID string) error {
	// 這裡應該包含更複雜的邏輯，如刪除相關的伺服器、訊息等
	// 暫時只刪除用戶記錄
	return us.userRepo.DeleteUser(userID)
}
