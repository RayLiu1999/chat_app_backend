# Mocks 目錄

這個目錄包含了測試中使用的共用 mock 實作，讓所有測試可以重複使用這些 mock 物件。

## 檔案列表

### 1. `odm_mock.go`
- **目的**: Mock `providers.ODM` 介面
- **用途**: 用於測試需要資料庫操作的服務
- **主要方法**:
  - `FindOne()` - 查詢單一文檔
  - `Find()` - 查詢多個文檔
  - `Create()` - 創建文檔
  - `Update()` - 更新文檔
  - `Delete()` - 刪除文檔
  - 以及其他資料庫操作方法

### 2. `server_member_repository_mock.go`
- **目的**: Mock `repositories.ServerMemberRepository` 介面
- **用途**: 用於測試伺服器成員相關功能
- **主要方法**:
  - `IsMemberOfServer()` - 檢查用戶是否為伺服器成員
  - `AddMemberToServer()` - 將用戶添加到伺服器
  - `RemoveMemberFromServer()` - 從伺服器移除用戶
  - `GetServerMembers()` - 獲取伺服器所有成員
  - `GetUserServers()` - 獲取用戶加入的所有伺服器
  - `UpdateMemberRole()` - 更新成員角色
  - `GetMemberCount()` - 獲取伺服器成員數量

## 使用範例

### 使用 ODM Mock

```go
import (
    "chat_app_backend/app/mocks"
    "chat_app_backend/app/models"
    "github.com/stretchr/testify/mock"
    "testing"
)

func TestSomeFunction(t *testing.T) {
    // 創建 mock 實例
    mockODM := new(mocks.ODM)
    
    // 設定期望的行為
    mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.User")).
        Return(models.User{Username: "testuser"}, nil).Once()
    
    // 使用 mock 進行測試
    // ...
    
    // 驗證所有期望的方法都被調用了
    mockODM.AssertExpectations(t)
}
```

### 使用 ServerMemberRepository Mock

```go
import (
    "chat_app_backend/app/mocks"
    "github.com/stretchr/testify/mock"
    "testing"
)

func TestCheckMembership(t *testing.T) {
    // 創建 mock 實例
    mockRepo := new(mocks.ServerMemberRepository)
    
    // 設定期望的行為
    mockRepo.On("IsMemberOfServer", "server123", "user456").
        Return(true, nil).Once()
    
    // 使用 mock 進行測試
    isMember, err := mockRepo.IsMemberOfServer("server123", "user456")
    
    // 斷言結果
    assert.True(t, isMember)
    assert.NoError(t, err)
    
    // 驗證所有期望的方法都被調用了
    mockRepo.AssertExpectations(t)
}
```

## 設計原則

1. **單一職責**: 每個 mock 檔案對應一個介面
2. **完整實作**: Mock 必須實作介面的所有方法
3. **使用 testify/mock**: 使用 stretchr/testify 的 mock 框架來簡化測試
4. **返回值正確**: 確保返回值的型別和數量與原始介面一致

## 新增 Mock 的步驟

1. 在 `app/mocks/` 目錄下創建新檔案，命名格式為 `<interface_name>_mock.go`
2. 定義 Mock 結構體，嵌入 `mock.Mock`
3. 實作介面的所有方法，使用 `m.Called()` 來記錄和返回預期的值
4. 在測試中引用 `chat_app_backend/app/mocks` 套件來使用 mock

## 注意事項

- Mock 的方法簽名必須與原始介面完全一致
- 使用 `mock.Anything` 來匹配任何參數值
- 使用 `mock.AnythingOfType("*models.XXX")` 來匹配特定型別
- 記得在測試結束時調用 `AssertExpectations(t)` 來驗證所有期望的方法都被調用
- Mock 返回值必須包含所有返回參數（包括錯誤）

## 參考資料

- [testify/mock 文檔](https://pkg.go.dev/github.com/stretchr/testify/mock)
- [Go Testing 最佳實踐](https://golang.org/doc/effective_go#testing)
