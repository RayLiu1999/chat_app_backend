package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCSRFToken(t *testing.T) {
	token, err := GenerateCSRFToken()

	assert.NoError(t, err, "GenerateCSRFToken should not return an error")
	assert.NotEmpty(t, token, "Generated token should not be empty")

	// 檢查 token 是否為合法的 base64 raw url 編碼字串
	// 這只是簡單檢查，更嚴謹的做法應該要解碼驗證。
	assert.Regexp(t, `^[A-Za-z0-9_-]+$`, token, "Token 應為合法的 base64 raw url 編碼")

	// 再產生一個 token，確保兩者不同
	token2, err2 := GenerateCSRFToken()
	assert.NoError(t, err2)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token, token2, "兩次產生的 token 不應相同")
}
