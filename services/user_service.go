package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"chat_app_backend/repositories"
	"chat_app_backend/utils"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	config       *config.Config
	mongoConnect *mongo.Database
	userRepo     repositories.UserRepositoryInterface
	odm          *providers.ODM
}

func NewUserService(cfg *config.Config, mongodb *mongo.Database, userRepo repositories.UserRepositoryInterface) *UserService {
	return &UserService{
		config:       cfg,
		mongoConnect: mongodb,
		userRepo:     userRepo,
		odm:          providers.NewODM(mongodb),
	}
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
func (us *UserService) GetUserResponseById(objectID primitive.ObjectID) (*models.UserResponse, error) {
	user, err := us.userRepo.GetUserById(objectID)
	if err != nil {
		return nil, err
	}

	// 轉換為 UserResponse
	response := &models.UserResponse{
		ID:       objectID.Hex(),
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Picture:  user.Picture,
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
			Displayable: true,
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

	err = us.odm.InsertOne(context.Background(), &refreshTokenDoc)
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
	_, objectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		return &utils.AppError{
			Code: utils.ErrUnauthorized,
		}
	}

	// 使用 UpdateMany 直接更新所有符合條件的文檔
	filter := bson.M{"user_id": objectID, "revoked": false}
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
			Displayable: true,
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
			Displayable: true,
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
