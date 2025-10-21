package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD5Hash(t *testing.T) {
	input := "hello world"
	expected := "5eb63bbbe01eeed093cb22bb8f5acdc3"
	assert.Equal(t, expected, MD5Hash(input))
}

func TestSHA256Hash(t *testing.T) {
	input := "hello world"
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	assert.Equal(t, expected, SHA256Hash(input))
}

func TestGenerateHMAC(t *testing.T) {
	message := "test message"
	key := "secret_key"

	t.Run("MD5", func(t *testing.T) {
		expected := "486fa85728f963477fd4101fde7d9c82"
		hmac, err := GenerateHMAC(message, key, "md5")
		assert.NoError(t, err)
		assert.Equal(t, expected, hmac)
	})

	t.Run("SHA1", func(t *testing.T) {
		expected := "4148fadb734189bfc41220c36e1039e4e81e0ffd"
		hmac, err := GenerateHMAC(message, key, "sha1")
		assert.NoError(t, err)
		assert.Equal(t, expected, hmac)
	})

	t.Run("SHA256", func(t *testing.T) {
		expected := "bfbd2f9aaf48aec78dc00cd1c36433f95648b5ad13a5d9d15c9e148b4b882a1a"
		hmac, err := GenerateHMAC(message, key, "sha256")
		assert.NoError(t, err)
		assert.Equal(t, expected, hmac)
	})

	t.Run("不支援的演算法", func(t *testing.T) {
		_, err := GenerateHMAC(message, key, "sha512")
		assert.Error(t, err)
	})
}

func TestAESEncryptDecrypt(t *testing.T) {
	key := []byte("a very secret key 12345678901234") // 32 bytes for AES-256
	plaintext := []byte("this is a super secret message")

	t.Run("成功加密和解密", func(t *testing.T) {
		encrypted, err := AESEncrypt(plaintext, key)
		assert.NoError(t, err)
		assert.NotEmpty(t, encrypted)

		decrypted, err := AESDecrypt(encrypted, key)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("無效的金鑰", func(t *testing.T) {
		invalidKey := []byte("short")
		_, err := AESEncrypt(plaintext, invalidKey)
		assert.Error(t, err, "使用無效金鑰大小進行加密應該失敗")

		// 假設我們有一個有效的加密字符串
		encrypted, _ := AESEncrypt(plaintext, key)
		_, err = AESDecrypt(encrypted, invalidKey)
		assert.Error(t, err, "使用無效金鑰大小進行解密應該失敗")
	})

	t.Run("損毀的資料", func(t *testing.T) {
		encrypted, _ := AESEncrypt(plaintext, key)
		corrupted := "corrupted" + encrypted
		_, err := AESDecrypt(corrupted, key)
		assert.Error(t, err, "解密損毀資料應該失敗")
	})
}
