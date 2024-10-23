# API 清單

## 認證

### POST /register
- **描述**: 用於用戶註冊
- **請求參數**:
  - `username` (string): 用戶名
  - `password` (string): 密碼
- **回應**: 返回是否註冊成功

### POST /login
- **描述**: 用於用戶登錄
- **請求參數**:
  - `username` (string): 用戶名
  - `password` (string): 密碼
- **回應**: 返回用戶的 access token，並把 refresh token 設置在 cookie 中

### POST /logout
- **描述**: 用於用戶登出
- **回應**: 無

### POST /refresh_token
- **描述**: 用於刷新 access token
- **回應**: 返回用戶的新 access token

## 用戶

### GET /user
- **描述**: 獲取用戶資訊
- **Header**:
  - `Authorization`: 用戶的 access token
- **回應**: 返回用戶資訊

### GET /friends
- **描述**: 獲取用戶的好友清單
- **Header**:
  - `Authorization`: 用戶的 access token
- **回應**: 返回用戶的好友清單

## 聊天

### WebSocket /ws
- **描述**: 用於聊天
- **請求參數**:
  - `token` (string): 用戶的 access token
- **回應**: 返回聊天訊息

## GET /servers
- **描述**: 獲取用戶加入的伺服器清單
- **Header**:
  - `Authorization`: 用戶的 access token
- **回應**: 返回用戶加入伺服器的清單

## GET /channels/:server_id/:channel_id
- **描述**: 獲取用戶在伺服器底下的聊天頻道
- **請求參數**:
  - `server_id` (int): 伺服器 ID
  - `channel_id` (int): 頻道 ID
- **Header**:
  - `Authorization`: 用戶的 access token
- **回應**: 返回所有聊天頻道的清單

## GET /chat_rooms
- **描述**: 獲取用戶加入的聊天室清單
- **Header**:
  - `Authorization`: 用戶的 access token
- **回應**: 返回用戶加入聊天室的清單