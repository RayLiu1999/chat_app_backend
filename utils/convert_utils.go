package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StructToJSON 將結構體轉換為 JSON 字串
// 參數：
//   - v: 要轉換的結構體
//   - pretty: 是否美化輸出
//
// 返回：
//   - JSON 字串和錯誤信息
func StructToJSON(v any, pretty ...bool) (string, error) {
	var (
		bytes []byte
		err   error
	)

	if len(pretty) > 0 && pretty[0] {
		bytes, err = json.MarshalIndent(v, "", "  ")
	} else {
		bytes, err = json.Marshal(v)
	}

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// JSONToStruct 解析 JSON 字串到結構體
// 參數：
//   - jsonStr: JSON 字串
//   - v: 結構體指針
//
// 返回：
//   - 錯誤信息
func JSONToStruct(jsonStr string, v any) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// StructToMap 將結構體轉換為 map[string]any
// 參數：
//   - obj: 要轉換的結構體
//
// 返回：
//   - 轉換後的 map 和錯誤信息
func StructToMap(obj any) (map[string]any, error) {
	// 先轉為 JSON
	jsonStr, err := StructToJSON(obj)
	if err != nil {
		return nil, err
	}

	// 再解析為 map
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// DeepCopy 使用 JSON 序列化/反序列化進行深度複製
// 參數：
//   - src: 源物件
//   - dst: 目標物件（必須是指針）
//
// 返回：
//   - 錯誤信息
func DeepCopy(src, dst any) error {
	if src == nil {
		return errors.New("源物件不能為 nil")
	}

	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, dst)
}

// ToStr 將任意類型安全地轉換為字串
// 參數：
//   - value: 任意類型值
//
// 返回：
//   - 轉換後的字串
func ToStr(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339Nano)
	case []byte:
		return string(v)
	default:
		// 嘗試使用 JSON 轉換
		jsonStr, err := StructToJSON(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return jsonStr
	}
}

// ToObjectID 嘗試將字串轉換為 ObjectID，若格式錯誤回傳零值與 false
// 主要是為了統一處理 primitive.ObjectIDFromHex 的錯誤，避免直接拋出底層 Error
func ToObjectID(id string) (primitive.ObjectID, bool) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, false
	}
	return objID, true
}

// MustObjectID 嘗試將字串轉換為 ObjectID，若格式錯誤回傳零值 (用於已驗證過格式的場景)
func MustObjectID(id string) primitive.ObjectID {
	objID, _ := primitive.ObjectIDFromHex(id)
	return objID
}
