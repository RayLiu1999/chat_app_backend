package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsString(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		str   string
		want  bool
	}{
		{"字符串存在", []string{"a", "b", "c"}, "b", true},
		{"字符串不存在", []string{"a", "b", "c"}, "d", false},
		{"空切片", []string{}, "a", false},
		{"包含空字符串的切片", []string{"", "b"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ContainsString(tt.slice, tt.str))
		})
	}
}

func TestContainsInt(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		val   int
		want  bool
	}{
		{"整數存在", []int{1, 2, 3}, 2, true},
		{"整數不存在", []int{1, 2, 3}, 4, false},
		{"空切片", []int{}, 1, false},
		{"包含零的切片", []int{0, 1, 2}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ContainsInt(tt.slice, tt.val))
		})
	}
}

func TestRemoveString(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		str   string
		want  []string
	}{
		{"移除現有元素", []string{"a", "b", "c"}, "b", []string{"a", "c"}},
		{"移除不存在的元素", []string{"a", "b", "c"}, "d", []string{"a", "b", "c"}},
		{"從包含重複的切片中移除", []string{"a", "b", "b", "c"}, "b", []string{"a", "c"}},
		{"從空切片中移除", []string{}, "a", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, RemoveString(tt.slice, tt.str))
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		want  []string
	}{
		{"包含重複元素的切片", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"沒有重複元素的切片", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"空切片", []string{}, []string{}},
		{"包含空字符串的切片", []string{"a", "", "b", ""}, []string{"a", "", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UniqueStrings(tt.slice)
			// 注意: 結果中元素的順序不一定與 `want` 相同
			// 如果實現改變。對於當前實現，它是穩定的。
			assert.Equal(t, tt.want, got)
		})
	}
}
