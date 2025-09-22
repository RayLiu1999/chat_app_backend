package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
)

// PrettyPrint 將任何資料結構轉換為美觀的 JSON 格式並打印到終端
// prefix: 可選的前綴說明文字
// data: 任何要打印的資料結構
func PrettyPrint(prefix string, data interface{}) {
	// 獲取調用者信息
	pc, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	parts := strings.Split(funcName, ".")
	callerFunc := parts[len(parts)-1]
	parts = strings.Split(file, "/")
	callerFile := parts[len(parts)-1]

	// 構建標頭
	header := fmt.Sprintf("\033[1;36m[DEBUG] %s:%d %s()\033[0m", callerFile, line, callerFunc)

	// 如果提供了前綴，添加它
	if prefix != "" {
		header += fmt.Sprintf(" \033[1;33m%s\033[0m", prefix)
	}

	// 打印標頭
	log.Println(header)

	// 處理 nil 值
	if data == nil {
		log.Println("\033[1;31m<nil>\033[0m")
		return
	}

	// 判斷是否為空值
	v := reflect.ValueOf(data)
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
		log.Println("\033[1;31m<nil>\033[0m")
		return
	}

	// 轉換為 JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("\033[1;31m無法轉換為 JSON: %v\033[0m\n原始數據: %+v\n", err, data)
		return
	}

	// 打印 JSON 數據
	log.Printf("\033[1;32m%s\033[0m\n", string(jsonData))
}

// PrettyPrintf 打印格式化的調試信息
// format: 格式化字符串
// args: 變長參數列表
func PrettyPrintf(format string, args ...interface{}) {
	// 獲取調用者信息
	pc, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	parts := strings.Split(funcName, ".")
	callerFunc := parts[len(parts)-1]
	parts = strings.Split(file, "/")
	callerFile := parts[len(parts)-1]

	// 構建標頭
	header := fmt.Sprintf("\033[1;36m[DEBUG] %s:%d %s()\033[0m", callerFile, line, callerFunc)

	// 上色格式化字符串
	coloredFormat := "\033[1;33m" + format + "\033[0m"

	log.Printf("%s: "+coloredFormat, append([]interface{}{header}, args...)...)
}

// PrettyPrintError 打印錯誤信息
func PrettyPrintError(prefix string, err error) {
	if err == nil {
		return
	}

	pc, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	parts := strings.Split(funcName, ".")
	callerFunc := parts[len(parts)-1]
	parts = strings.Split(file, "/")
	callerFile := parts[len(parts)-1]

	header := fmt.Sprintf("\033[1;31m[ERROR] %s:%d %s()\033[0m", callerFile, line, callerFunc)

	if prefix != "" {
		header += fmt.Sprintf(" \033[1;33m%s\033[0m", prefix)
	}

	log.Printf("%s: %v\n", header, err)
}
