package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateCSRFToken 產生一個隨機的 CSRF token
func GenerateCSRFToken() (string, error) {
	// 產生 32 bytes 的隨機數據
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %v", err)
	}

	// 將隨機數據編碼為 base64 字串（使用 RawURLEncoding 避免 = 填充）
	token := base64.RawURLEncoding.EncodeToString(bytes)
	return token, nil
}
