package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"chat_app_backend/config"
)

// SimpleGet 發送簡單的 GET 請求並返回響應內容
// 參數：
//   - url: 請求 URL
//   - headers: 可選的請求頭
//
// 返回：
//   - 響應內容和錯誤信息
func SimpleGet(url string, headers ...map[string]string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 添加請求頭
	if len(headers) > 0 {
		for key, value := range headers[0] {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 請求失敗，狀態碼: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// PostJSON 發送 JSON 格式的 POST 請求
// 參數：
//   - url: 請求 URL
//   - data: 請求數據（會被轉換為 JSON）
//   - headers: 可選的請求頭
//
// 返回：
//   - 響應內容和錯誤信息
func PostJSON(url string, data interface{}, headers ...map[string]string) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// 設置 Content-Type
	req.Header.Set("Content-Type", "application/json")

	// 添加其他請求頭
	if len(headers) > 0 {
		for key, value := range headers[0] {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// DownloadFile 下載檔案並保存到本地
// 參數：
//   - url: 檔案 URL
//   - filepath: 保存路徑
//
// 返回：
//   - 錯誤信息
func DownloadFile(url, filepath string) error {
	// 創建目錄
	dir := path.Dir(filepath)
	if err := EnsureDir(dir); err != nil {
		return err
	}

	// 創建檔案
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 發送 GET 請求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 檢查響應狀態
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下載失敗，HTTP 狀態碼: %d", resp.StatusCode)
	}

	// 複製內容到檔案
	_, err = io.Copy(out, resp.Body)
	return err
}

// NewHTTPClient 創建一個配置了超時和代理的 HTTP 客戶端
// 參數：
//   - timeout: 超時時間（秒）
//   - proxyURL: 代理 URL，可選
//
// 返回：
//   - HTTP 客戶端
func NewHTTPClient(timeout int, proxyURL ...string) (*http.Client, error) {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// 設置代理
	if len(proxyURL) > 0 && proxyURL[0] != "" {
		proxyURLParsed, err := url.Parse(proxyURL[0])
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURLParsed)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	return client, nil
}

// ParseURLParams 解析 URL 查詢參數為 map
// 參數：
//   - urlStr: URL 字串
//
// 返回：
//   - 參數 map 和錯誤信息
func ParseURLParams(urlStr string) (map[string]string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	params := make(map[string]string)
	for key, values := range parsedURL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	return params, nil
}

// 取的upload的url
// 參數：
//   - path: 路徑
//   - params: 參數 map
//
// 返回：
//   - URL 字串
func GetUploadURL(path string, params map[string]string) string {
	cfg := config.GetConfig()
	baseURL := cfg.Server.BaseURL
	uploadPath := "/uploads/"
	url := baseURL + uploadPath + path

	if len(params) > 0 {
		paramStr := "?"
		for key, value := range params {
			paramStr += key + "=" + value + "&"
		}
		url += paramStr[:len(paramStr)-1]
	}

	return url
}
