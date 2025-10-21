package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeAtoi(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultVal int
		expected   int
	}{
		{"有效的整數", "123", 0, 123},
		{"無效的整數", "abc", 10, 10},
		{"空字符串", "", 5, 5},
		{"負整數", "-50", 0, -50},
		{"零", "0", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, SafeAtoi(tt.input, tt.defaultVal))
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		min      int
		max      int
		expected int
	}{
		{"值在範圍內", 5, 0, 10, 5},
		{"值低於最小值", -5, 0, 10, 0},
		{"值高於最大值", 15, 0, 10, 10},
		{"值在最小值", 0, 0, 10, 0},
		{"值在最大值", 10, 0, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Clamp(tt.val, tt.min, tt.max))
		})
	}
}

func TestRoundFloat(t *testing.T) {
	tests := []struct {
		name      string
		val       float64
		precision int
		expected  float64
	}{
		{"四捨五入", 1.235, 2, 1.24},
		{"四捨五入向下", 1.234, 2, 1.23},
		{"不需要四捨五入", 1.23, 2, 1.23},
		{"零精度", 1.5, 0, 2.0},
		{"負數", -1.235, 2, -1.24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, RoundFloat(tt.val, tt.precision))
		})
	}
}

func TestIsInRange(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		min      int
		max      int
		expected bool
	}{
		{"值在範圍內", 5, 0, 10, true},
		{"值低於最小值", -1, 0, 10, false},
		{"值高於最大值", 11, 0, 10, false},
		{"值在最小值", 0, 0, 10, true},
		{"值在最大值", 10, 0, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsInRange(tt.val, tt.min, tt.max))
		})
	}
}

func TestRandomInt(t *testing.T) {
	t.Run("最小值等於最大值", func(t *testing.T) {
		min, max := 10, 10
		assert.Equal(t, min, RandomInt(min, max))
	})

	t.Run("最小值大於最大值", func(t *testing.T) {
		min, max := 20, 10
		assert.Equal(t, min, RandomInt(min, max))
	})

	t.Run("有效範圍", func(t *testing.T) {
		min, max := 1, 100
		for i := 0; i < 1000; i++ {
			got := RandomInt(min, max)
			assert.GreaterOrEqual(t, got, min)
			assert.LessOrEqual(t, got, max)
		}
	})
}
