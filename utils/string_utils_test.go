package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeSubstring(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		start    int
		end      int
		expected string
	}{
		{"正常情況", "hello world", 0, 5, "hello"},
		{"開始超出邊界（負數）", "hello world", -2, 5, "hello"},
		{"結束超出邊界（正數）", "hello world", 6, 20, "world"},
		{"開始大於結束", "hello world", 5, 0, ""},
		{"空字符串", "", 0, 5, ""},
		{"完整字符串", "abc", 0, 3, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, SafeSubstring(tt.s, tt.start, tt.end))
		})
	}
}

func TestIsEmptyOrWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected bool
	}{
		{"空字符串", "", true},
		{"僅空格", "   ", true},
		{"僅制表符", "\t\t", true},
		{"混合空白符", " \n\t ", true},
		{"包含內容的字符串", "hello", false},
		{"前導/尾隨空白的字符串", "  hello  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsEmptyOrWhitespace(tt.s))
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected string
	}{
		{"正常情況", "hello", "Hello"},
		{"已大寫", "World", "World"},
		{"單一字符", "a", "A"},
		{"空字符串", "", ""},
		{"非字母開頭", "1world", "1world"},
		{"多位元組字符", "你好", "你好"}, // unicode.ToUpper 處理這個
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Capitalize(tt.s))
		})
	}
}

func TestMaskString(t *testing.T) {
	tests := []struct {
		name              string
		s                 string
		visibleStartChars int
		visibleEndChars   int
		maskChar          []rune
		expected          string
	}{
		{"正常情況", "1234567890", 3, 4, nil, "123***7890"},
		{"短字符串", "abc", 2, 2, nil, "abc"},
		{"自定義遮蔽字符", "1234567890", 2, 2, []rune{'#'}, "12######90"},
		{"空字符串", "", 2, 2, nil, ""},
		{"零個可見字符", "123456", 0, 0, nil, "******"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, MaskString(tt.s, tt.visibleStartChars, tt.visibleEndChars, tt.maskChar...))
		})
	}
}

func TestRandomString(t *testing.T) {
	t.Run("預設字符集", func(t *testing.T) {
		length := 10
		s := RandomString(length)
		assert.Equal(t, length, len(s))
		// 檢查它是否只包含字母數字字符
		for _, r := range s {
			assert.Contains(t, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", string(r))
		}
	})

	t.Run("自定義字符集", func(t *testing.T) {
		length := 12
		charset := "abc"
		s := RandomString(length, charset)
		assert.Equal(t, length, len(s))
		for _, r := range s {
			assert.Contains(t, charset, string(r))
		}
	})

	t.Run("零長度", func(t *testing.T) {
		s := RandomString(0)
		assert.Equal(t, "", s)
	})
}
