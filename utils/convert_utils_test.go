package utils

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestStructToJSON(t *testing.T) {
	s := testStruct{Name: "Alice", Age: 30}

	t.Run("正常轉換", func(t *testing.T) {
		jsonStr, err := StructToJSON(s)
		assert.NoError(t, err, "StructToJSON() 不應返回錯誤")

		var result testStruct
		err = json.Unmarshal([]byte(jsonStr), &result)
		assert.NoError(t, err, "解析產生的 JSON 失敗")

		assert.Equal(t, s, result)
	})

	t.Run("美化格式", func(t *testing.T) {
		jsonStr, err := StructToJSON(s, true)
		assert.NoError(t, err, "StructToJSON() 美化格式不應返回錯誤")

		// 預期美化格式：以 "{\n" 開頭，以 "\n}" 結尾，包含 "  "
		assert.True(t, jsonStr[0] == '{' && jsonStr[1] == '\n' && jsonStr[len(jsonStr)-1] == '}', "StructToJSON() 美化格式不正確。取得:\n%s", jsonStr)
	})
}

func TestJSONToStruct(t *testing.T) {
	jsonStr := `{"name":"Bob","age":40}`
	var result testStruct
	expected := testStruct{Name: "Bob", Age: 40}

	err := JSONToStruct(jsonStr, &result)
	assert.NoError(t, err, "JSONToStruct() should not return an error")
	assert.Equal(t, expected, result)

	t.Run("無效 JSON", func(t *testing.T) {
		invalidJson := `{name:"Bob"}`
		var r testStruct
		err := JSONToStruct(invalidJson, &r)
		assert.Error(t, err, "JSONToStruct() 使用無效 JSON 應返回錯誤")
	})
}

func TestStructToMap(t *testing.T) {
	s := testStruct{Name: "Charlie", Age: 25}
	expected := map[string]any{"name": "Charlie", "age": float64(25)} // JSON 將數字解析為 float64

	m, err := StructToMap(s)
	assert.NoError(t, err, "StructToMap() 不應返回錯誤")
	assert.Equal(t, expected, m)
}

func TestDeepCopy(t *testing.T) {
	src := testStruct{Name: "David", Age: 50}
	var dst testStruct

	err := DeepCopy(src, &dst)
	assert.NoError(t, err, "DeepCopy() 不應返回錯誤")
	assert.Equal(t, src, dst)

	// 修改原始值，目標值不應改變
	src.Age = 51
	assert.NotEqual(t, src.Age, dst.Age, "DeepCopy() 沒有執行深拷貝，目標值被修改了")

	t.Run("空值來源", func(t *testing.T) {
		var d testStruct
		err := DeepCopy(nil, &d)
		assert.Error(t, err, "DeepCopy() 使用空值來源應返回錯誤")
	})
}

func TestToStr(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"String", "hello", "hello"},
		{"Int", 123, "123"},
		{"Float", 3.14, "3.14"},
		{"Bool true", true, "true"},
		{"Bool false", false, "false"},
		{"Nil", nil, ""},
		{"Struct", testStruct{Name: "Eve", Age: 28}, `{"name":"Eve","age":28}`},
		{"Time", now, now.Format(time.RFC3339Nano)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToStr(tt.value)
			// 針對結構體，我們需要比較未序列化的物件
			if tt.name == "Struct" {
				assert.JSONEq(t, tt.expected, got, "ToStr() 對結構體應產生等效的 JSON")
			} else {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
