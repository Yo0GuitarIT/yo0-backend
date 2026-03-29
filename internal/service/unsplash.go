package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Yo0GuitarIT/yo0-backend/internal/config"
	"github.com/Yo0GuitarIT/yo0-backend/internal/model"
)

// GetRandomPhoto 向 Unsplash API 請求一張隨機照片
// 回傳：照片資料、HTTP 狀態碼、錯誤
func GetRandomPhoto() (*model.Photo, int, error) {
	accessKey := config.UnsplashAccessKey()

	req, err := http.NewRequest("GET", "https://api.unsplash.com/photos/random", nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// Unsplash 使用 Client-ID 作為認證方式
	req.Header.Set("Authorization", fmt.Sprintf("Client-ID %s", accessKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var photo model.Photo
	if err := json.Unmarshal(body, &photo); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &photo, resp.StatusCode, nil
}
