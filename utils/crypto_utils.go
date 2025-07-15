package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"
)

// MD5Hash 計算字串的 MD5 哈希值
// 參數：
//   - s: 輸入字串
//
// 返回：
//   - MD5 哈希值（16 進制字串）
func MD5Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// SHA256Hash 計算字串的 SHA-256 哈希值
// 參數：
//   - s: 輸入字串
//
// 返回：
//   - SHA-256 哈希值（16 進制字串）
func SHA256Hash(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// GenerateHMAC 使用指定算法生成 HMAC
// 參數：
//   - message: 消息內容
//   - key: 密鑰
//   - algorithm: 哈希算法，支持 "md5", "sha1", "sha256"
//
// 返回：
//   - HMAC 哈希值（16 進制字串）和錯誤信息
func GenerateHMAC(message, key, algorithm string) (string, error) {
	var mac hash.Hash

	switch strings.ToLower(algorithm) {
	case "md5":
		mac = hmac.New(md5.New, []byte(key))
	case "sha1":
		mac = hmac.New(sha1.New, []byte(key))
	case "sha256":
		mac = hmac.New(sha256.New, []byte(key))
	default:
		return "", fmt.Errorf("不支持的哈希算法: %s", algorithm)
	}

	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// AESEncrypt 使用 AES 算法加密數據
// 參數：
//   - plaintext: 明文數據
//   - key: 密鑰（16、24 或 32 位元組，對應 AES-128、AES-192 或 AES-256）
//
// 返回：
//   - 加密後的 base64 編碼字串和錯誤信息
func AESEncrypt(plaintext, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 創建初始向量
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// 加密
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	copy(ciphertext[:aes.BlockSize], iv)
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// 轉換為 base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt 使用 AES 算法解密數據
// 參數：
//   - encryptedData: 加密後的 base64 字串
//   - key: 解密密鑰
//
// 返回：
//   - 解密後的明文和錯誤信息
func AESDecrypt(encryptedData string, key []byte) ([]byte, error) {
	// 解碼 base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	// 檢查密文長度
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("密文太短")
	}

	// 從密文中提取初始向量
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// 創建解密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 解密
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

// GenerateSecureRandomString 生成一個指定長度的安全隨機字串
// 參數：
//   - length: 字串長度
//
// 返回：
//   - 隨機字串和錯誤信息
func GenerateSecureRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	// 使用 crypto/rand 而不是 math/rand
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// 映射到字元集
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b), nil
}
