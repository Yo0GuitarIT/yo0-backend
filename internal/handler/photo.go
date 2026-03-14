package handler

import (
	"github.com/Yo0GuitarIT/yo0-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// GetRandomPhoto 處理 GET /photos/random 請求
// 呼叫 service 層取得照片資料後，直接回傳 Unsplash 的原始狀態碼與 JSON
func GetRandomPhoto(c *gin.Context) {
	result, statusCode, err := service.GetRandomPhoto()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
