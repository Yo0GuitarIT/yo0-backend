package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// GetRandomPhoto 向 Unsplash API 請求一張隨機照片
// 回傳：照片資料、HTTP 狀態碼、錯誤
func GetRandomPhoto() (map[string]interface{}, int, error) {
	// 從環境變數讀取 API Key
	accessKey := os.Getenv("UNSPLASH_ACCESS_KEY")

	req, err := http.NewRequest("GET", "https://api.unsplash.com/photos/random", nil)
	if err != nil {
		return nil, 500, err
	}
	// Unsplash 使用 Client-ID 作為認證方式
	req.Header.Set("Authorization", fmt.Sprintf("Client-ID %s", accessKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 500, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 500, err
	}

	// 將 JSON response 解析成 map
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 500, err
	}

	return result, resp.StatusCode, nil
}
