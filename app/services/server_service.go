package services

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"chat_app_backend/app/repositories"
	"chat_app_backend/config"
	"chat_app_backend/utils"
	"context"
	"fmt"
	"mime/multipart"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServerService struct {
	config              *config.Config
	serverRepo          repositories.ServerRepositoryInterface
	serverMemberRepo    repositories.ServerMemberRepositoryInterface
	userRepo            repositories.UserRepositoryInterface
	channelRepo         repositories.ChannelRepositoryInterface
	channelCategoryRepo repositories.ChannelCategoryRepositoryInterface
	chatRepo            repositories.ChatRepositoryInterface
	odm                 *providers.ODM
	fileUploadService   FileUploadServiceInterface
	userService         UserServiceInterface
}

func NewServerService(cfg *config.Config, odm *providers.ODM, serverRepo repositories.ServerRepositoryInterface, serverMemberRepo repositories.ServerMemberRepositoryInterface, userRepo repositories.UserRepositoryInterface, channelRepo repositories.ChannelRepositoryInterface, channelCategoryRepo repositories.ChannelCategoryRepositoryInterface, chatRepo repositories.ChatRepositoryInterface, fileUploadService FileUploadServiceInterface, userService UserServiceInterface) *ServerService {
	return &ServerService{
		config:              cfg,
		serverRepo:          serverRepo,
		serverMemberRepo:    serverMemberRepo,
		userRepo:            userRepo,
		channelRepo:         channelRepo,
		channelCategoryRepo: channelCategoryRepo,
		chatRepo:            chatRepo,
		odm:                 odm,
		fileUploadService:   fileUploadService,
		userService:         userService,
	}
}

// UpdateUserService 更新 UserService 引用
func (ss *ServerService) UpdateUserService(userService UserServiceInterface) {
	ss.userService = userService
}

// getUserPictureURL 獲取用戶頭像 URL（從 ObjectID 解析）
func (ss *ServerService) getUserPictureURL(user *models.User) string {
	if user.PictureID.IsZero() || ss.fileUploadService == nil {
		return ""
	}

	pictureURL, err := ss.fileUploadService.GetFileURLByID(user.PictureID.Hex())
	if err != nil {
		return ""
	}
	return pictureURL
}

// GetServerListResponse 獲取用戶的伺服器列表回應格式
func (ss *ServerService) GetServerListResponse(userID string) ([]models.ServerResponse, *models.MessageOptions) {
	// 驗證用戶是否存在
	_, err := ss.userRepo.GetUserById(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrUserNotFound,
			Details: err,
			Message: "用戶不存在",
		}
	}

	// 獲取用戶的伺服器列表
	serverMembers, err := ss.serverMemberRepo.GetUserServers(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "獲取伺服器列表失敗",
		}
	}

	serverIDs := make([]primitive.ObjectID, len(serverMembers))
	for i, member := range serverMembers {
		serverIDs[i] = member.ServerID
	}

	// 獲取伺服器詳細信息
	var servers []models.Server
	err = ss.odm.Find(context.Background(), bson.M{"_id": bson.M{"$in": serverIDs}}, &servers)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "獲取伺服器詳細信息失敗",
		}
	}

	// 轉換為響應格式
	serverResponses := make([]models.ServerResponse, len(servers))
	for i, server := range servers {
		var pictureURL string
		// 如果有圖片ID，則取得圖片URL
		if !server.ImageID.IsZero() {
			if url, err := ss.fileUploadService.GetFileURLByID(server.ImageID.Hex()); err == nil {
				pictureURL = url
			}
		}

		serverResponses[i] = models.ServerResponse{
			ID:          server.BaseModel.GetID(),
			Name:        server.Name,
			PictureURL:  pictureURL,
			Description: server.Description,
		}
	}

	return serverResponses, nil
}

// CreateServer 創建伺服器
func (ss *ServerService) CreateServer(userID string, name string, file multipart.File, header *multipart.FileHeader) (*models.ServerResponse, *models.MessageOptions) {
	// 驗證用戶是否存在
	_, err := ss.userRepo.GetUserById(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrUserNotFound,
			Details: err,
			Message: "用戶不存在",
		}
	}

	// 將string userID轉換為ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Details: err,
			Message: "無效的用戶ID格式",
		}
	}

	// 上傳檔案 - 使用伺服器專用配置
	uploadResult := &models.FileResult{}
	if file != nil {
		serverConfig := models.GetServerUploadConfig()
		result, msgOpt := ss.fileUploadService.UploadFileWithConfig(file, header, userID, serverConfig)
		if msgOpt != nil {
			return nil, &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Details: msgOpt.Details,
				Message: "伺服器圖片上傳失敗: " + msgOpt.Message,
			}
		}
		uploadResult = result
	}

	// 創建伺服器
	server := &models.Server{
		Name:        name,
		ImageID:     uploadResult.ID, // 使用上傳後的圖片名稱
		Description: "This is a test server",
		OwnerID:     userObjectID,
		MemberCount: 1,          // 初始只有擁有者
		IsPublic:    true,       // 預設設為公開，可以在後續版本中讓用戶選擇
		Tags:        []string{}, // 可以在未來版本中添加標籤功能
		Region:      "TW",       // 預設地區為台灣
		MaxMembers:  100,        // 預設最大成員數
	}

	// 保存到資料庫
	createdServer, err := ss.serverRepo.CreateServer(server)
	if err != nil {
		// 如果保存失敗，嘗試刪除已上傳的檔案
		if deleteErr := ss.fileUploadService.DeleteFile(uploadResult.FilePath); deleteErr != nil {
			fmt.Printf("清理上傳檔案失敗: %v\n", deleteErr)
		}
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "創建伺服器失敗",
		}
	}

	utils.PrettyPrint("Created Server", createdServer)

	// 將擁有者添加到 ServerMember 表
	err = ss.serverMemberRepo.AddMemberToServer(createdServer.BaseModel.GetID().Hex(), userID, "owner")
	if err != nil {
		fmt.Printf("添加擁有者到成員列表失敗: %v\n", err)
	}

	// 創建預設頻道
	err = ss.createDefaultChannels(createdServer.BaseModel.GetID())
	if err != nil {
		// 如果創建頻道失敗，記錄錯誤但不阻止伺服器創建
		fmt.Printf("創建預設頻道失敗: %v\n", err)
	}

	// 返回響應格式
	var pictureURL string
	// 如果有圖片ID，則取得圖片URL
	if !createdServer.ImageID.IsZero() {
		if url, err := ss.fileUploadService.GetFileURLByID(createdServer.ImageID.Hex()); err == nil {
			pictureURL = url
		}
	}

	serverResponse := &models.ServerResponse{
		ID:          createdServer.BaseModel.GetID(),
		Name:        createdServer.Name,
		PictureURL:  pictureURL,
		Description: createdServer.Description,
	}

	return serverResponse, nil
}

// createDefaultChannels 為新創建的伺服器創建預設頻道和類別
func (ss *ServerService) createDefaultChannels(serverID primitive.ObjectID) error {
	// 首先創建預設類別
	err := ss.createDefaultCategories(serverID)
	if err != nil {
		return fmt.Errorf("創建預設類別失敗: %v", err)
	}

	// 獲取剛創建的預設類別
	textCategory, voiceCategory, err := ss.getDefaultCategoriesByServerID(serverID.Hex())
	if err != nil {
		return fmt.Errorf("獲取預設類別失敗: %v", err)
	}

	// 創建預設文字頻道
	textChannel := &models.Channel{
		Name:       "一般",
		ServerID:   serverID,
		CategoryID: textCategory.BaseModel.GetID(),
		Type:       "text",
	}

	err = ss.channelRepo.CreateChannel(textChannel)
	if err != nil {
		return fmt.Errorf("創建預設文字頻道失敗: %v", err)
	}

	// 創建預設語音頻道
	voiceChannel := &models.Channel{
		Name:       "語音聊天室",
		ServerID:   serverID,
		CategoryID: voiceCategory.BaseModel.GetID(),
		Type:       "voice",
	}

	err = ss.channelRepo.CreateChannel(voiceChannel)
	if err != nil {
		return fmt.Errorf("創建預設語音頻道失敗: %v", err)
	}

	return nil
}

// createDefaultCategories 創建預設類別
func (ss *ServerService) createDefaultCategories(serverID primitive.ObjectID) error {
	// 創建文字類別
	textCategory := &models.ChannelCategory{
		Name:         "文字頻道",
		ServerID:     serverID,
		CategoryType: "text",
		Position:     1,
	}

	err := ss.channelCategoryRepo.CreateChannelCategory(textCategory)
	if err != nil {
		return fmt.Errorf("創建文字類別失敗: %v", err)
	}

	// 創建語音類別
	voiceCategory := &models.ChannelCategory{
		Name:         "語音頻道",
		ServerID:     serverID,
		CategoryType: "voice",
		Position:     2,
	}

	err = ss.channelCategoryRepo.CreateChannelCategory(voiceCategory)
	if err != nil {
		return fmt.Errorf("創建語音類別失敗: %v", err)
	}

	return nil
}

// getDefaultCategoriesByServerID 獲取伺服器的預設類別（文字和語音）
func (ss *ServerService) getDefaultCategoriesByServerID(serverID string) (*models.ChannelCategory, *models.ChannelCategory, error) {
	categories, err := ss.channelCategoryRepo.GetChannelCategoriesByServerID(serverID)
	if err != nil {
		return nil, nil, err
	}

	var textCategory, voiceCategory *models.ChannelCategory

	for i := range categories {
		switch categories[i].CategoryType {
		case "text":
			textCategory = &categories[i]
		case "voice":
			voiceCategory = &categories[i]
		}
	}

	if textCategory == nil || voiceCategory == nil {
		return nil, nil, fmt.Errorf("找不到預設類別")
	}

	return textCategory, voiceCategory, nil
}

// SearchPublicServers 搜尋公開伺服器
func (ss *ServerService) SearchPublicServers(userID string, request models.ServerSearchRequest) (*models.ServerSearchResults, *models.MessageOptions) {
	// 驗證用戶是否存在
	_, err := ss.userRepo.GetUserById(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "用戶不存在",
			Details: err.Error(),
		}
	}

	// 執行搜尋
	servers, totalCount, err := ss.serverRepo.SearchPublicServers(request)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "搜尋伺服器失敗",
			Details: err.Error(),
		}
	}

	// 轉換為響應格式
	serverResponses := make([]models.ServerSearchResponse, 0, len(servers))
	for _, server := range servers {
		// 檢查用戶是否已加入伺服器
		isJoined, err := ss.serverMemberRepo.IsMemberOfServer(server.BaseModel.GetID().Hex(), userID)
		if err != nil {
			isJoined = false // 出錯時假設未加入
		}

		// 獲取伺服器擁有者信息
		var ownerName string
		if owner, err := ss.userRepo.GetUserById(server.OwnerID.Hex()); err == nil {
			ownerName = owner.Username
		}

		// 獲取圖片URL
		var pictureURL string
		if !server.ImageID.IsZero() {
			if url, err := ss.fileUploadService.GetFileURLByID(server.ImageID.Hex()); err == nil {
				pictureURL = url
			}
		}

		serverResponse := models.ServerSearchResponse{
			ID:          server.BaseModel.GetID(),
			Name:        server.Name,
			PictureURL:  pictureURL,
			Description: server.Description,
			MemberCount: server.MemberCount,
			IsJoined:    isJoined,
			OwnerName:   ownerName,
			CreatedAt:   server.CreatedAt.UnixMilli(),
		}

		serverResponses = append(serverResponses, serverResponse)
	}

	// 計算總頁數
	totalPages := int(totalCount) / request.Limit
	if int(totalCount)%request.Limit > 0 {
		totalPages++
	}

	return &models.ServerSearchResults{
		Servers:    serverResponses,
		TotalCount: totalCount,
		Page:       request.Page,
		Limit:      request.Limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateServer 更新伺服器信息
func (ss *ServerService) UpdateServer(userID string, serverID string, updates map[string]interface{}) (*models.ServerResponse, *models.MessageOptions) {
	// 驗證用戶是否存在
	_, err := ss.userRepo.GetUserById(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "用戶不存在",
			Details: err.Error(),
		}
	}

	// 獲取伺服器信息
	server, err := ss.serverRepo.GetServerByID(serverID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "伺服器不存在",
			Details: err.Error(),
		}
	}

	// 檢查用戶是否有權限更新（擁有者或管理員）
	if server.OwnerID.Hex() != userID {
		// 可以在未來添加管理員權限檢查
		return nil, &models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "無權限更新此伺服器",
		}
	}

	// 執行更新
	err = ss.serverRepo.UpdateServer(serverID, updates)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "更新伺服器失敗",
			Details: err.Error(),
		}
	}

	// 獲取更新後的伺服器信息
	updatedServer, err := ss.serverRepo.GetServerByID(serverID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取更新後伺服器信息失敗",
			Details: err.Error(),
		}
	}

	// 獲取圖片URL
	var pictureURL string
	if !updatedServer.ImageID.IsZero() {
		if url, err := ss.fileUploadService.GetFileURLByID(updatedServer.ImageID.Hex()); err == nil {
			pictureURL = url
		}
	}

	return &models.ServerResponse{
		ID:          updatedServer.BaseModel.GetID(),
		Name:        updatedServer.Name,
		PictureURL:  pictureURL,
		Description: updatedServer.Description,
	}, nil
}

// DeleteServer 刪除伺服器
func (ss *ServerService) DeleteServer(userID string, serverID string) *models.MessageOptions {
	// 驗證用戶是否存在
	_, err := ss.userRepo.GetUserById(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "用戶不存在",
			Details: err.Error(),
		}
	}

	// 獲取伺服器信息
	server, err := ss.serverRepo.GetServerByID(serverID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "伺服器不存在",
			Details: err.Error(),
		}
	}

	// 檢查用戶是否為擁有者
	if server.OwnerID.Hex() != userID {
		return &models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "只有伺服器擁有者可以刪除伺服器",
		}
	}

	// 刪除所有相關的伺服器成員記錄
	members, _, err := ss.serverMemberRepo.GetServerMembers(serverID, 1, 1000)
	if err == nil {
		for _, member := range members {
			ss.serverMemberRepo.RemoveMemberFromServer(serverID, member.UserID.Hex())
		}
	}

	// 刪除所有相關的頻道和類別
	err = ss.deleteServerChannelsAndCategories(serverID)
	if err != nil {
		fmt.Printf("刪除伺服器頻道和類別失敗: %v\n", err)
		// 不阻止伺服器刪除，只記錄錯誤
	}

	// 刪除伺服器圖片
	if !server.ImageID.IsZero() {
		if msgOpt := ss.fileUploadService.DeleteFileByID(server.ImageID.Hex(), userID); msgOpt != nil {
			fmt.Printf("刪除伺服器圖片失敗: %v\n", msgOpt.Details)
		}
	}

	// 最後刪除伺服器
	err = ss.serverRepo.DeleteServer(serverID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "刪除伺服器失敗",
			Details: err.Error(),
		}
	}

	return nil
}

// GetServerByID 根據ID獲取伺服器信息
func (ss *ServerService) GetServerByID(userID string, serverID string) (*models.ServerResponse, *models.MessageOptions) {
	// 驗證用戶是否存在
	_, err := ss.userRepo.GetUserById(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "用戶不存在",
			Details: err.Error(),
		}
	}

	// 檢查用戶是否有權限查看此伺服器（是成員或伺服器是公開的）
	isMember, err := ss.serverMemberRepo.IsMemberOfServer(serverID, userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "檢查成員身份失敗",
			Details: err.Error(),
		}
	}

	// 獲取伺服器信息
	server, err := ss.serverRepo.GetServerByID(serverID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "伺服器不存在",
			Details: err.Error(),
		}
	}

	// 如果不是成員且伺服器不公開，則無權限查看
	if !isMember && !server.IsPublic {
		return nil, &models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "無權限查看此伺服器",
		}
	}

	// 獲取圖片URL
	var pictureURL string
	if !server.ImageID.IsZero() {
		if url, err := ss.fileUploadService.GetFileURLByID(server.ImageID.Hex()); err == nil {
			pictureURL = url
		}
	}

	return &models.ServerResponse{
		ID:          server.BaseModel.GetID(),
		Name:        server.Name,
		PictureURL:  pictureURL,
		Description: server.Description,
	}, nil
}

// GetServerDetailByID 獲取伺服器詳細信息（包含成員和頻道列表）
func (ss *ServerService) GetServerDetailByID(userID string, serverID string) (*models.ServerDetailResponse, *models.MessageOptions) {
	// 驗證用戶是否存在
	_, err := ss.userRepo.GetUserById(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "用戶不存在",
			Details: err.Error(),
		}
	}

	// 檢查用戶是否有權限查看此伺服器（是成員或伺服器是公開的）
	isMember, err := ss.serverMemberRepo.IsMemberOfServer(serverID, userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "檢查成員身份失敗",
			Details: err.Error(),
		}
	}

	// 獲取伺服器信息
	server, err := ss.serverRepo.GetServerByID(serverID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "伺服器不存在",
			Details: err.Error(),
		}
	}

	// 如果不是成員且伺服器不公開，則無權限查看詳細信息
	if !isMember && !server.IsPublic {
		return nil, &models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "無權限查看此伺服器",
		}
	}

	// 獲取圖片URL
	var pictureURL string
	if !server.ImageID.IsZero() {
		if url, err := ss.fileUploadService.GetFileURLByID(server.ImageID.Hex()); err == nil {
			pictureURL = url
		}
	}

	// 獲取成員列表（只有成員才能看到完整成員列表）
	var members []models.ServerMemberResponse
	if isMember {
		serverMembers, _, err := ss.serverMemberRepo.GetServerMembers(serverID, 1, 100)
		if err != nil {
			return nil, &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "獲取成員列表失敗",
				Details: err.Error(),
			}
		}

		// 獲取所有成員的用戶信息
		var userIDs []string
		for _, member := range serverMembers {
			userIDs = append(userIDs, member.UserID.Hex())
		}

		users, err := ss.userRepo.GetUserListByIds(userIDs)
		if err != nil {
			return nil, &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "獲取用戶信息失敗",
				Details: err.Error(),
			}
		}

		// 創建用戶信息映射
		userMap := make(map[string]models.User)
		for _, user := range users {
			userMap[user.ID.Hex()] = user
		}

		// 組合成員響應
		for _, member := range serverMembers {
			if user, exists := userMap[member.UserID.Hex()]; exists {
				// 使用伺服器內暱稱，如果沒有則使用用戶暱稱
				displayNickname := member.Nickname
				if displayNickname == "" {
					displayNickname = user.Nickname
				}

				// 檢查用戶在線狀態
				isOnline := false
				if ss.userService != nil {
					isOnline = ss.userService.IsUserOnlineByWebSocket(member.UserID.Hex())
				}

				members = append(members, models.ServerMemberResponse{
					UserID:       member.UserID.Hex(),
					Username:     user.Username,
					Nickname:     displayNickname,
					PictureURL:   ss.getUserPictureURL(&user),
					Role:         member.Role,
					IsOnline:     isOnline,
					LastActiveAt: member.LastActiveAt.UnixMilli(),
					JoinedAt:     member.JoinedAt.UnixMilli(),
				})
			}
		}
	}

	// 獲取頻道列表（只有成員才能看到頻道）
	var channels []models.ChannelResponse
	if isMember {
		serverChannels, err := ss.channelRepo.GetChannelsByServerID(serverID)
		if err != nil {
			return nil, &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "獲取頻道列表失敗",
				Details: err.Error(),
			}
		}

		for _, channel := range serverChannels {
			channels = append(channels, models.ChannelResponse{
				ID:       channel.BaseModel.GetID(),
				ServerID: channel.ServerID,
				Name:     channel.Name,
				Type:     channel.Type,
			})
		}
	}

	return &models.ServerDetailResponse{
		ID:          server.BaseModel.GetID(),
		Name:        server.Name,
		PictureURL:  pictureURL,
		Description: server.Description,
		MemberCount: server.MemberCount,
		IsPublic:    server.IsPublic,
		OwnerID:     server.OwnerID.Hex(),
		Members:     members,
		Channels:    channels,
	}, nil
}

// GetServerChannels 獲取伺服器的頻道列表
func (ss *ServerService) GetServerChannels(serverID string) ([]models.Channel, *models.MessageOptions) {
	channels, err := ss.channelRepo.GetChannelsByServerID(serverID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取頻道列表失敗",
			Details: err.Error(),
		}
	}
	return channels, nil
}

// JoinServer 請求加入伺服器
func (ss *ServerService) JoinServer(userID string, serverID string) *models.MessageOptions {
	// 驗證用戶是否存在
	_, err := ss.userRepo.GetUserById(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "用戶不存在",
			Details: err.Error(),
		}
	}

	// 獲取伺服器信息
	server, err := ss.serverRepo.GetServerByID(serverID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "伺服器不存在",
			Details: err.Error(),
		}
	}

	// 檢查伺服器是否為公開伺服器
	if !server.IsPublic {
		return &models.MessageOptions{
			Code:    models.ErrForbidden,
			Message: "此伺服器不開放加入",
		}
	}

	// 檢查用戶是否已經是成員
	isMember, err := ss.serverMemberRepo.IsMemberOfServer(serverID, userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "檢查成員身份失敗",
			Details: err.Error(),
		}
	}
	if isMember {
		return &models.MessageOptions{
			Code:    models.ErrOperationFailed,
			Message: "您已經是此伺服器的成員",
		}
	}

	// 檢查伺服器是否已達到最大成員數限制
	memberCount, err := ss.serverMemberRepo.GetMemberCount(serverID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取成員數量失敗",
			Details: err.Error(),
		}
	}
	if int(memberCount) >= server.MaxMembers {
		return &models.MessageOptions{
			Code:    models.ErrForbidden,
			Message: "伺服器已達到最大成員數限制",
		}
	}

	// 添加用戶到伺服器（默認角色為 member）
	err = ss.serverMemberRepo.AddMemberToServer(serverID, userID, "member")
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "加入伺服器失敗",
			Details: err.Error(),
		}
	}

	// 更新伺服器成員數量快取
	newMemberCount := int(memberCount) + 1
	err = ss.serverRepo.UpdateMemberCount(serverID, newMemberCount)
	if err != nil {
		fmt.Printf("更新成員數量快取失敗: %v\n", err)
	}

	return nil
}

// LeaveServer 離開伺服器
func (ss *ServerService) LeaveServer(userID string, serverID string) *models.MessageOptions {
	// 驗證用戶是否存在
	_, err := ss.userRepo.GetUserById(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "用戶不存在",
			Details: err.Error(),
		}
	}

	// 獲取伺服器信息
	server, err := ss.serverRepo.GetServerByID(serverID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrNotFound,
			Message: "伺服器不存在",
			Details: err.Error(),
		}
	}

	// 檢查用戶是否為伺服器擁有者
	if server.OwnerID.Hex() == userID {
		return &models.MessageOptions{
			Code:    models.ErrForbidden,
			Message: "伺服器擁有者無法離開伺服器，請先轉移擁有權或刪除伺服器",
		}
	}

	// 檢查用戶是否為成員
	isMember, err := ss.serverMemberRepo.IsMemberOfServer(serverID, userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "檢查成員身份失敗",
			Details: err.Error(),
		}
	}
	if !isMember {
		return &models.MessageOptions{
			Code:    models.ErrOperationFailed,
			Message: "您不是此伺服器的成員",
		}
	}

	// 從伺服器移除用戶
	err = ss.serverMemberRepo.RemoveMemberFromServer(serverID, userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "離開伺服器失敗",
			Details: err.Error(),
		}
	}

	// 更新伺服器成員數量快取
	memberCount, err := ss.serverMemberRepo.GetMemberCount(serverID)
	if err == nil {
		err = ss.serverRepo.UpdateMemberCount(serverID, int(memberCount))
		if err != nil {
			fmt.Printf("更新成員數量快取失敗: %v\n", err)
		}
	}

	return nil
}

// deleteServerChannelsAndCategories 刪除伺服器的所有頻道和類別
func (ss *ServerService) deleteServerChannelsAndCategories(serverID string) error {
	// 獲取伺服器的所有頻道
	channels, err := ss.channelRepo.GetChannelsByServerID(serverID)
	if err != nil {
		return fmt.Errorf("獲取伺服器頻道失敗: %v", err)
	}

	// 刪除所有頻道及其訊息
	for _, channel := range channels {
		// 先刪除該頻道的所有訊息
		err = ss.chatRepo.DeleteMessagesByRoomID(channel.BaseModel.GetID().Hex())
		if err != nil {
			fmt.Printf("刪除頻道 %s 的訊息失敗: %v\n", channel.Name, err)
		}

		// 然後刪除頻道本身
		err = ss.channelRepo.DeleteChannel(channel.BaseModel.GetID().Hex())
		if err != nil {
			// 記錄錯誤但繼續刪除其他頻道
			fmt.Printf("刪除頻道 %s 失敗: %v\n", channel.Name, err)
		}
	}

	// 獲取伺服器的所有類別
	categories, err := ss.channelCategoryRepo.GetChannelCategoriesByServerID(serverID)
	if err != nil {
		return fmt.Errorf("獲取伺服器類別失敗: %v", err)
	}

	// 刪除所有類別
	for _, category := range categories {
		err = ss.channelCategoryRepo.DeleteChannelCategory(category.BaseModel.GetID().Hex())
		if err != nil {
			// 記錄錯誤但繼續刪除其他類別
			fmt.Printf("刪除類別 %s 失敗: %v\n", category.Name, err)
		}
	}

	return nil
}
