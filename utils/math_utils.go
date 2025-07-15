package utils

import (
	"math"
	"math/rand"
	"strconv"
)

// SafeAtoi 安全地將字串轉換為整數，避免異常
// 參數：
//   - s: 要轉換的字串
//   - defaultVal: 轉換失敗時返回的默認值
//
// 返回：
//   - 轉換後的整數
func SafeAtoi(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}

	return val
}

// Clamp 將數值限制在指定範圍內
// 參數：
//   - val: 要限制的值
//   - min: 最小值
//   - max: 最大值
//
// 返回：
//   - 限制後的值
func Clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// RoundFloat 將浮點數四捨五入到指定小數位
// 參數：
//   - val: 要四捨五入的浮點數
//   - precision: 小數位數
//
// 返回：
//   - 四捨五入後的浮點數
func RoundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// IsInRange 檢查整數是否在指定範圍內
// 參數：
//   - val: 要檢查的值
//   - min: 範圍最小值
//   - max: 範圍最大值
//
// 返回：
//   - 是否在範圍內
func IsInRange(val, min, max int) bool {
	return val >= min && val <= max
}

// RandomInt 生成指定範圍內的隨機整數
// 參數：
//   - min: 最小值（包含）
//   - max: 最大值（包含）
//
// 返回：
//   - 隨機整數
func RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	return rand.Intn(max-min+1) + min
}
