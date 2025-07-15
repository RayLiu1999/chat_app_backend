package utils

import (
	"math/rand"
	"strings"
	"unicode"
)

// SafeSubstring 安全地截取字串，避免索引越界
// 參數：
//   - s: 原始字串
//   - start: 起始索引（包含）
//   - end: 結束索引（不包含）
//
// 返回：
//   - 截取後的字串
func SafeSubstring(s string, start, end int) string {
	if start < 0 {
		start = 0
	}

	length := len(s)
	if end > length {
		end = length
	}

	if start > end {
		return ""
	}

	return s[start:end]
}

// IsEmptyOrWhitespace 檢查字串是否為空或僅包含空白字元
// 參數：
//   - s: 待檢查的字串
//
// 返回：
//   - 若字串為空或僅包含空白字元，則返回 true；否則返回 false
func IsEmptyOrWhitespace(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// Capitalize 將字串的首字母轉為大寫，其餘保持不變
// 參數：
//   - s: 原始字串
//
// 返回：
//   - 首字母大寫的字串
func Capitalize(s string) string {
	if s == "" {
		return ""
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// MaskString 對敏感資訊進行脫敏處理，如信用卡號、手機號等
// 參數：
//   - s: 原始字串
//   - visibleStartChars: 開頭可見字元數
//   - visibleEndChars: 結尾可見字元數
//   - maskChar: 用於替換的字元，默認為 '*'
//
// 返回：
//   - 脫敏後的字串
func MaskString(s string, visibleStartChars, visibleEndChars int, maskChar ...rune) string {
	if s == "" {
		return ""
	}

	length := len(s)
	if visibleStartChars+visibleEndChars >= length {
		return s
	}

	mask := '*'
	if len(maskChar) > 0 {
		mask = maskChar[0]
	}

	result := []rune(s)
	for i := visibleStartChars; i < length-visibleEndChars; i++ {
		result[i] = mask
	}

	return string(result)
}

// RandomString 生成指定長度的隨機字串
// 參數：
//   - length: 字串長度
//   - charset: 可選的字元集，默認為字母和數字
//
// 返回：
//   - 隨機生成的字串
func RandomString(length int, charset ...string) string {
	var chars string
	if len(charset) > 0 {
		chars = charset[0]
	} else {
		chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}
