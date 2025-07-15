package utils

// ContainsString 檢查字符串切片中是否包含指定字符串
// 參數：
//   - slice: 要檢查的字符串切片
//   - str: 要查找的字符串
//
// 返回：
//   - 如果切片中包含該字符串，則返回 true；否則返回 false
func ContainsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// ContainsInt 檢查整數切片中是否包含指定整數
// 參數：
//   - slice: 要檢查的整數切片
//   - val: 要查找的整數
//
// 返回：
//   - 如果切片中包含該整數，則返回 true；否則返回 false
func ContainsInt(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// RemoveString 從字符串切片中移除指定字符串（保持順序）
// 參數：
//   - slice: 原始字符串切片
//   - str: 要移除的字符串
//
// 返回：
//   - 移除指定字符串後的新切片
func RemoveString(slice []string, str string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != str {
			result = append(result, item)
		}
	}
	return result
}

// UniqueStrings 返回字符串切片中的唯一元素（去重）
// 參數：
//   - slice: 原始字符串切片
//
// 返回：
//   - 去重後的字符串切片
func UniqueStrings(slice []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
